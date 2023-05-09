package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
)

type Client struct {
	httpClient *http.Client
	url        string
}

var ErrMakeRequest = errors.New("error creating request")

func New(host, port string) *Client {
	return &Client{
		url:        fmt.Sprintf("http://%s:%s/api/v1", host, port),
		httpClient: new(http.Client),
	}
}

func (c *Client) Create(ctx context.Context, clientID string, chatID int64) error {
	clientIDCasted, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return fmt.Errorf("primitive.ObjectIDFromHex: %w", err)
	}

	body := entities.TGAccount{
		ClientID: clientIDCasted,
		ChatID:   chatID,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	res, err := c.httpClient.Post(c.url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("http.Post: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("%w: %d", ErrMakeRequest, res.StatusCode)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, clientID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", c.url, clientID), http.NoBody)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", ErrMakeRequest, res.StatusCode)
	}

	return nil
}
