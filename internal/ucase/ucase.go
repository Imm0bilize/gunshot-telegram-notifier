package ucase

import (
	"context"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
)

type (
	ClientUseCase interface {
		Create(ctx context.Context, account entities.TGAccount) error
		Delete(ctx context.Context, clientID string) error
	}

	NotificationUseCase interface {
		Notify(ctx context.Context, message entities.NotificationMessage) error
	}
)

type UCase struct {
	ClientUCase       ClientUseCase
	NotificationUCase NotificationUseCase
}

func NewUCase(client ClientUseCase, notification NotificationUseCase) *UCase {
	return &UCase{
		ClientUCase:       client,
		NotificationUCase: notification,
	}
}
