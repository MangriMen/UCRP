package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
	token   string
}

type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) Get(ctx context.Context, path string, query url.Values) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, query, nil, nil)
}

func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, nil, payload, map[string]string{"Content-Type": "application/json"})
}

func (c *Client) Put(ctx context.Context, path string, body any) (*Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPut, path, nil, payload, map[string]string{"Content-Type": "application/json"})
}

func (c *Client) Patch(ctx context.Context, path string, body any) (*Response, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	return c.do(ctx, http.MethodPatch, path, nil, payload, map[string]string{"Content-Type": "application/json"})
}

func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string, fileName string, fileContent []byte) (*Response, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("write multipart field %q: %w", key, err)
		}
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("create multipart file part: %w", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("write multipart file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}
	return c.do(ctx, http.MethodPost, path, nil, buf.Bytes(), map[string]string{"Content-Type": writer.FormDataContentType()})
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte, headers map[string]string) (*Response, error) {
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, fmt.Errorf("%s %s: build request: %w", method, path, err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s %s: read body: %w", method, path, err)
	}
	return &Response{Status: resp.StatusCode, Header: resp.Header, Body: raw}, nil
}

func (r *Response) Decode(v any) error {
	if len(r.Body) == 0 {
		return fmt.Errorf("empty response body")
	}
	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("decode response body: %w (body=%s)", err, r.Snippet())
	}
	return nil
}

func (r *Response) JSON() (any, error) {
	if len(r.Body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}
	var v any
	if err := json.Unmarshal(r.Body, &v); err != nil {
		return nil, fmt.Errorf("decode response body: %w (body=%s)", err, r.Snippet())
	}
	return v, nil
}

func (r *Response) Snippet() string {
	const limit = 512
	s := strings.TrimSpace(string(r.Body))
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

type ErrorBody struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
		Details   []struct {
			Field  string `json:"field"`
			Reason string `json:"reason"`
		} `json:"details"`
	} `json:"error"`
}

func (r *Response) ErrorBody() (ErrorBody, bool) {
	var body ErrorBody
	if err := json.Unmarshal(r.Body, &body); err != nil {
		return body, false
	}
	if body.Error.Code == "" {
		return body, false
	}
	return body, true
}

func (r *Response) ErrorCode() string {
	body, ok := r.ErrorBody()
	if !ok {
		return ""
	}
	return body.Error.Code
}

func (r *Response) RequestID() string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	body, ok := r.ErrorBody()
	if !ok {
		return ""
	}
	return body.Error.RequestID
}
