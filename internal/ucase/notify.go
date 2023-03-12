package ucase

import (
	"context"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type (
	ChatIDGetter interface {
		GetChatIDByClientID(ctx context.Context, clientID string) (int64, error)
	}

	ClientNotifier interface {
		NotifyClient(ctx context.Context, chatID int64, msg entities.NotificationMessage) error
	}
)

type Notify struct {
	repo     ChatIDGetter
	notifier ClientNotifier
	tracer   trace.Tracer
}

func NewNotifyUCase(repo ChatIDGetter, notifier ClientNotifier) *Notify {
	return &Notify{
		repo:     repo,
		notifier: notifier,
		tracer:   otel.Tracer("notifyUCase"),
	}
}

func (n Notify) Notify(ctx context.Context, msg entities.NotificationMessage) error {
	ctx, span := n.tracer.Start(ctx, "ClientRepo.Delete")
	defer span.End()

	chatID, err := n.repo.GetChatIDByClientID(ctx, msg.ClientID)
	if err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "repo.GetChatIDByClientID")
	}

	if err = n.notifier.NotifyClient(ctx, chatID, msg); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "notifier.NotifyClient")
	}

	return nil
}
