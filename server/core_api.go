package server

import (
	"context"
	"fmt"
	"novaserver/server/dto"

	"github.com/nova-chat/novaproto"
	"github.com/nova-chat/novaproto/dhellman"
)

func (srv *Server) dhSv(ctx context.Context, client *Client, header novaproto.PacketHeader, clientHello dto.DHPublic) error {
	serverPublic, err := dhellman.GenerateKeyPair()
	if err != nil {
		return err
	}
	sharedKey, err := serverPublic.ComputeShared(clientHello.PublicKey)
	if err != nil {
		return err
	}

	if err := Send(client, ServerDHPublic, dto.DHPublic{PublicKey: serverPublic.PublicKey()}); err != nil {
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

	return nil
}
