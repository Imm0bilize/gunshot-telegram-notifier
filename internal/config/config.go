package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type KafkaConsumerConfig struct {
	Peers string `env:"KAFKA_PEERS"`
	Group string `env:"KAFKA_GROUP"`
	Topic string `env:"KAFKA_TOPIC"`
}

type BotConfig struct {
	Token string `env:"BOT_TOKEN"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
}

type GRPCConfig struct {
	Port string `env:"GRPC_PORT"`
}

type OTELConfig struct {
	Host string `env:"OTEL_HOST"`
	Port string `env:"OTEL_PORT"`
}

type Config struct {
	Bot   BotConfig
	DB    DBConfig
	Kafka KafkaConsumerConfig
	OTEL  OTELConfig
	GRPC  GRPCConfig
}

func New(envFiles ...string) (*Config, error) {
	err := godotenv.Load(envFiles...)
	if err != nil {
		return nil, errors.Wrap(err, "godotenv.Load")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process")
	}

	return &cfg, nil
}
