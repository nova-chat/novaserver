package server

import (
	"context"
	"fmt"
	"novaserver/server/dto"

	"github.com/nova-chat/novaproto"
	"github.com/nova-chat/novaproto/dhellman"
	"github.com/nova-chat/novaproto/serializer"
)

func (srv *Server) dhSv(ctx context.Context, client *User, header novaproto.PacketHeader, clientHello dto.DHPublic) error {

	serverPublic, err := dhellman.GenerateKeyPair()
	if err != nil {
		return err
	}
	sharedKey, err := serverPublic.ComputeShared(clientHello.PublicKey)
	if err != nil {
		return err
	}

	serverMsg, err := serializer.Marshal(dhellman.NewHelloMessage(serverPublic))
	if err != nil {
		return err
	}
	packetStream, err := client.PacketStream.SendPacket(novaproto.PacketHeader{
		Kind: ServerDHPublic.Value(),
	})
	_, err = packetStream.Write(serverMsg)
	if err != nil {
		return err
	}

	spk := serverPublic.PublicKey()
	salt := append(append([]byte{}, clientHello.PublicKey[:]...), spk[:]...)
	key, err := dhellman.DeriveKey(sharedKey, salt, []byte("novaproto/dhellman"))
	if err != nil {
		return err
	}
	client.EncryptionKey = key
	fmt.Println(client.EncryptionKey)

	return packetStream.Close()

}
