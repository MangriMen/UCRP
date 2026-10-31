package client

import (
	"context"
	"fmt"
)

func (c *Client) Login(ctx context.Context, email, password string) (*Response, error) {
	return c.Post(ctx, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
}

func (c *Client) Me(ctx context.Context) (*Response, error) {
	return c.Get(ctx, "/api/v1/auth/me", nil)
}

func (c *Client) LoginToken(ctx context.Context, email, password string) (string, *Response, error) {
	resp, err := c.Login(ctx, email, password)
	if err != nil {
		return "", nil, err
	}
	var payload struct {
		AccessToken string `json:"accessToken"`
	}
	if err := resp.Decode(&payload); err != nil {
		return "", resp, err
	}
	token := payload.AccessToken
	if token == "" {
		return "", resp, fmt.Errorf("login response carries no access token: %s", resp.Snippet())
	}
	return token, resp, nil
}
