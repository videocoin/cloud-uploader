package downloader

import (
	"context"
)

type key int

const (
	gdriveKey key = 1
)

func NewContextWithGDriveKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, gdriveKey, key)
}

func GDriveKeyFromContext(ctx context.Context) string {
	key, _ := ctx.Value(gdriveKey).(string)
	return key
}
