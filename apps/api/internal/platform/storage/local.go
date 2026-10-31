package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Local{root: abs}, nil
}

func (l *Local) path(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	rel := filepath.Clean(filepath.FromSlash(key))
	if rel == "." || filepath.IsAbs(rel) || hasDotDotPrefix(rel) || rel == ".." {
		return "", fmt.Errorf("invalid key")
	}
	full := filepath.Join(l.root, rel)
	if !isInside(l.root, full) {
		return "", fmt.Errorf("invalid key")
	}
	return full, nil
}

func (l *Local) Save(ctx context.Context, key string, r io.Reader) error {
	path, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (l *Local) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := l.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (l *Local) Delete(ctx context.Context, key string) error {
	path, err := l.path(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func isInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !hasDotDotPrefix(rel)
}

func hasDotDotPrefix(rel string) bool {
	return len(rel) >= 3 && rel[0] == '.' && rel[1] == '.' && (rel[2] == '/' || rel[2] == filepath.Separator)
}
