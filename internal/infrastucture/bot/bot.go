package bot

import (
	"context"
	"fmt"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/config"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const _messageTemplate = "Attention!\n" +
	"The system has been triggered for message from %s"

type Bot struct {
	botAPI *tgbotapi.BotAPI
	tracer trace.Tracer
}

func NewBot(cfg config.BotConfig) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, errors.Wrap(err, "tgbotapi.NewBotAPI")
	}

	return &Bot{
		botAPI: botAPI,
		tracer: otel.GetTracerProvider().Tracer("bot"),
	}, nil
}

func (b Bot) NotifyClient(ctx context.Context, chatID int64, msg entities.NotificationMessage) error {
	ctx, span := b.tracer.Start(ctx, "bot.NotifyClient")
	defer span.End()

	message := tgbotapi.NewMessage(chatID, fmt.Sprintf(_messageTemplate, msg.Timestamp.String()))
	if _, err := b.botAPI.Send(message); err != nil {
		span.RecordError(err)
		return errors.Wrap(err, "botAPI.Send")
	}

	return nil
}
