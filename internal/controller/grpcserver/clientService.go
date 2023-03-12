package grpcserver

import (
	"context"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/ucase"
	api "github.com/Imm0bilize/gunshot-telegram-notifier/pkg/api/proto"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ClientService struct {
	api.ClientServiceServer
	domain *ucase.UCase
}

func NewClientService(domain *ucase.UCase) *ClientService {
	return &ClientService{domain: domain}
}

func (c ClientService) CreateClientV1(ctx context.Context, req *api.CreateClientRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetClientId())
	if err != nil {
		return nil, errors.Wrap(err, "primitive.ObjectIDFromHex")
	}

	if err = c.domain.ClientUCase.Create(ctx, entities.TGAccount{
		ClientID: id,
		ChatID:   req.ChatId,
	}); err != nil {
		return nil, errors.Wrap(err, "domain.ClientUCase.Create")
	}
	return &emptypb.Empty{}, nil
}
func (c ClientService) DeleteClientV1(ctx context.Context, req *api.DeleteClientRequest) (*emptypb.Empty, error) {
	if err := c.domain.ClientUCase.Delete(ctx, req.GetClientId()); err != nil {
		return nil, errors.Wrap(err, "domain.ClientUCase.Delete")
	}
	return &emptypb.Empty{}, nil
}
