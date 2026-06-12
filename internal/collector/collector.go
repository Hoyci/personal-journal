package collector

import (
	"context"

	"github.com/hoyci/personal-journal/internal/core"
)

type Collector interface {
	Fetch(ctx context.Context) ([]core.Article, error)
	Name() string
}
