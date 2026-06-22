package port

import (
	"context"
)

type distributedSingleflight interface {
	Do(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error, bool)
}

type SingleflightCoordinator interface {
	distributedSingleflight
}
