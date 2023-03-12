package ucase

import (
	"context"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/pkg/errors"
)

type ClientRepo interface {
	Create(ctx context.Context, account entities.TGAccount) error
	Delete(ctx context.Context, clientID string) error
}

type Client struct {
	repo ClientRepo
}

func NewClientUCase(repo ClientRepo) *Client {
	return &Client{
		repo: repo,
	}
}

func (c Client) Create(ctx context.Context, account entities.TGAccount) error {
	if err := c.repo.Create(ctx, account); err != nil {
		return errors.Wrap(err, "repo.Create")
	}

	return nil
}

func (c Client) Delete(ctx context.Context, clientID string) error {
	if err := c.repo.Delete(ctx, clientID); err != nil {
		return errors.Wrap(err, "repo.Delete")
	}

	return nil
}
