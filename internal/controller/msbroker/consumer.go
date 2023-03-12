package msbroker

import (
	"context"
	"encoding/json"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/config"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/ucase"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"strings"
	"time"
)

const _notificationMethod = "telegram"

type KafkaConsumer struct {
	logger *zap.Logger
	domain *ucase.UCase
	group  sarama.ConsumerGroup
	ready  chan struct{}
	topics []string
}

func NewKafkaConsumer(cfg config.KafkaConsumerConfig, domain *ucase.UCase, logger *zap.Logger) (*KafkaConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_3_0_0
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Session.Timeout = 20 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 6 * time.Second
	saramaConfig.Consumer.MaxProcessingTime = 3 * time.Second

	group, err := sarama.NewConsumerGroup(strings.Split(cfg.Peers, ","), cfg.Group, saramaConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating consumer group client")
	}

	return &KafkaConsumer{
		domain: domain,
		logger: logger.Named("kafka-consumer"),
		ready:  make(chan struct{}),
		group:  group,
		topics: strings.Split(cfg.Topic, ","),
	}, nil
}

func (k KafkaConsumer) Run(ctx context.Context) error {
	errChan := make(chan error)

	go func() {
		for {
			if err := k.group.Consume(ctx, k.topics, &k); err != nil {
				k.logger.Error("error from consumer", zap.Error(err))
				errChan <- err
			}

			if ctx.Err() != nil {
				return
			}

			k.ready = make(chan struct{})
		}
	}()

	<-k.ready

	select {
	case <-ctx.Done():
		k.logger.Info("terminating consume: context canceled")
	case err := <-errChan:
		return err
	}

	if err := k.group.Close(); err != nil {
		return errors.Wrap(err, "error closing consumer group")
	}

	return nil
}

func (k KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	k.logger.Debug("Setup")
	return nil
}

func (k KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	k.logger.Debug("Cleanup")
	return nil
}

func containsMethod(necessary string, methods []string) bool {
	for _, method := range methods {
		if method == necessary {
			return true
		}
	}

	return false
}

func (k KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			go func() {
				var (
					msg entities.NotificationMessage
					err error
				)

				if err = json.Unmarshal(message.Value, &msg); err != nil {
					k.logger.Error("json.Unmarshal error: ", zap.Error(err))
				}

				if !containsMethod(_notificationMethod, msg.NotificationMethods) {
					session.MarkMessage(message, "")
					return
				}

				ctx := otel.GetTextMapPropagator().Extract(
					context.Background(), otelsarama.NewConsumerMessageCarrier(message),
				)

				if err = k.domain.NotificationUCase.Notify(ctx, msg); err != nil {
					k.logger.Error("domain.Notify error",
						zap.String("requestID", msg.RequestID.String()),
						zap.Error(err),
					)
				}

				session.MarkMessage(message, "")
			}()
		case <-session.Context().Done():
			return nil
		}
	}
}
