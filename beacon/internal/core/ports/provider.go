package ports

import (
	"context"

	"beacon/internal/core/entity"
)

type ChallengeProvider interface {
	FetchDaily(ctx context.Context) (entity.DailyChallenge, error)
}
