package domain

import (
	"context"
	"io"
)

type BlobStore interface {
	Save(ctx context.Context, key string, r io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttlSeconds int)
	Delete(key string)
}
