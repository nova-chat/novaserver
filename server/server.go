package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/nova-chat/novaproto"
)

type Server struct {
	lis net.Listener

	users *userStore

	routes map[uint64]RawPacketHandler
}

func NewServer(addr string) (*Server, error) {
	srv := &Server{
		users:  newUserStore(),
		routes: make(map[uint64]RawPacketHandler),
	}
	var err error

	srv.lis, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	ClientDHPublic.ServerRegister(srv, srv.dhSv)

	return srv, nil
}

func (r *Server) addRoute(method uint64, handler RawPacketHandler) {
	if _, ex := r.routes[method]; ex {
		log.Fatalf("route: %d already exist", method)
	}
	r.routes[method] = handler
}

func (r *Server) GetHandler(header novaproto.PacketHeader) (RawPacketHandler, bool) {
	handler, ex := r.routes[header.Kind]
	return handler, ex
}

func (srv *Server) Run(ctx context.Context) {
	go func() {
		tcpListener := srv.lis.(*net.TCPListener)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			tcpListener.SetDeadline(time.Now().Add(time.Second * 5))
			conn, err := tcpListener.Accept()
			if err != nil {
				// Check if the error is a timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Println("Accept error:", err)
				<-time.After(time.Second * 5)
				continue
			}
			go srv.handleConnection(ctx, conn)
		}
	}()
}

func (srv *Server) handleConnection(ctx context.Context, client net.Conn) {
	user := &User{
		Id:            uuid.New(),
		EncryptionKey: nil,
	}
	srv.users.SetUser(user)
	defer srv.users.DelUser(user.Id)

	user.WireStream = novaproto.NewNovaWireStreamCipher(novaproto.NewNovaWireStream(client))
	user.PacketStream = novaproto.NewRoutedPacketStream(novaproto.NewPacketStream(user.WireStream))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		header, dataStream, err := user.PacketStream.ReceivePacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("failed to recv packet: %v", err)
			continue
		}
		err = srv.handleRequest(ctx, user, header, dataStream)
		if err != nil {
			log.Printf("failed to call method %d: %v", header.Kind, err)
			continue
		}
	}
}

func (srv *Server) handleRequest(ctx context.Context, client *User, header novaproto.PacketHeader, dataStream io.Reader) error {
	if IsValueMethodKind(header.Kind, KindClient2Server) {
		handler, ex := srv.GetHandler(header)
		if !ex {
			io.Copy(io.Discard, dataStream)
			return fmt.Errorf("handler for method: %x not found", header.Kind)
		}
		return handler(ctx, client, header, dataStream)
	} else if IsValueMethodKind(header.Kind, KindClient2Client) {
		// Find target user.
		target, ex := srv.users.GetUser(header.TargetID)
		if !ex {
			io.Copy(io.Discard, dataStream)
			return fmt.Errorf("target %s not found", header.TargetID.String())
		}
		// Relay data
		outStream, err := target.PacketStream.SendPacket(novaproto.PacketHeader{
			SourceID: client.Id,
			TargetID: target.Id,
			Kind:     header.Kind,
		})
		if err != nil {
			io.Copy(io.Discard, dataStream)
			return fmt.Errorf("failed to start relay packet: %v", err)
		}
		_, err = io.Copy(outStream, dataStream)
		return err
	} else {
		return fmt.Errorf("unknown method kind signature")
	}
}
