package ports

import (
	"context"

	"beacon/internal/core/entity"
)

type Notifier interface {
	Send(ctx context.Context, msg entity.Message) error
}
