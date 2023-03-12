package api

import (
	"context"
	api "github.com/Imm0bilize/gunshot-telegram-notifier/pkg/api/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TGNotificationServiceClient struct {
	client api.ClientServiceClient
	conn   *grpc.ClientConn
}

func NewTGNotificationServiceClient(address string) (*TGNotificationServiceClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "grpc.Dial")
	}

	return &TGNotificationServiceClient{
		client: api.NewClientServiceClient(conn),
		conn:   conn,
	}, nil
}

func (t TGNotificationServiceClient) Create(ctx context.Context, clientID string, chatID int64) error {
	_, err := t.client.CreateClientV1(ctx, &api.CreateClientRequest{
		ClientId: clientID,
		ChatId:   chatID,
	})
	if err != nil {
		return errors.Wrap(err, "client.CreateClientV1")
	}

	return nil
}

func (t TGNotificationServiceClient) Delete(ctx context.Context, clientID string) error {
	_, err := t.client.DeleteClientV1(ctx, &api.DeleteClientRequest{ClientId: clientID})
	if err != nil {
		return errors.Wrap(err, "client.DeleteClientV1")
	}

	return nil
}

func (t TGNotificationServiceClient) Shutdown(_ context.Context) error {
	if err := t.conn.Close(); err != nil {
		return errors.Wrap(err, "conn.Close")
	}

	return nil
}
