package app

import (
	"context"
	"fmt"
	"log"
	"net"
	stdHttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/config"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/controller/http"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/controller/msbroker"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/infrastucture/bot"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/infrastucture/repository"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/ucase"
)

func createTraceProvider(cfg config.OTELConfig) func(context.Context) error {
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(net.JoinHostPort(cfg.Host, cfg.Port)),
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "gunshot-telegram-notifier"),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: ", err)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	return exporter.Shutdown
}

func createDB(cfg config.DBConfig) (*mongo.Database, func(context.Context) error, error) {
	clientOptions := options.Client()
	clientOptions.Monitor = otelmongo.NewMonitor()
	clientOptions.ApplyURI(fmt.Sprintf("mongodb://%s:%s/", cfg.Host, cfg.Port))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, errors.Wrap(err, "mongo.Connect")
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, nil, errors.Wrap(err, "client.Ping")
	}

	disconnect := func(ctx2 context.Context) error {
		return client.Disconnect(ctx)
	}

	return client.Database(cfg.Name), disconnect, nil
}

func Run(cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	shutdownTrace := createTraceProvider(cfg.OTEL)

	db, dbDisconnect, err := createDB(cfg.DB)
	if err != nil {
		logger.Fatal("can't create db connection", zap.Error(err))
	}

	repo := repository.NewRepository(db)
	notifierBot, err := bot.NewBot(cfg.Bot)
	if err != nil {
		logger.Fatal("can't create bot", zap.Error(err))
	}

	uCase := ucase.NewUCase(
		ucase.NewClientUCase(repo),
		ucase.NewNotifyUCase(repo, notifierBot),
	)

	broker, err := msbroker.NewKafkaConsumer(cfg.Kafka, uCase, logger)
	if err != nil {
		logger.Fatal("can't create kafka consumer", zap.Error(err))
	}

	go func() {
		if err := broker.Run(ctx); err != nil {
			logger.Fatal("error running consumer", zap.Error(err))
		}
	}()

	handler := http.NewHTTPServer(logger, uCase)
	server := stdHttp.Server{
		Addr:    net.JoinHostPort("", cfg.GRPC.Port),
		Handler: handler,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, stdHttp.ErrServerClosed) {
			logger.Fatal("error running http server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	ctx, shutdownFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownFunc()

	if err = server.Shutdown(ctx); err != nil {
		logger.Error("error shutting down http server", zap.Error(err))
	}

	if err = dbDisconnect(ctx); err != nil {
		logger.Error("error disconnecting db", zap.Error(err))
	}
	if err = shutdownTrace(ctx); err != nil {
		logger.Error("error shutting down tracing")
	}
}
