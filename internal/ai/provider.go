package ai

import (
	"context"
)

type ArticleInput struct {
	Index   int
	Title   string
	Summary string
	Source  string
}

type ClassificationResult struct {
	Index     int
	Relevant  bool
	Score     float64
	Reasoning string
}

type DayBriefing struct {
	Summary    string
	Highlights []string
}

type LLMProvider interface {
	ClassifyArticles(ctx context.Context, category string, articles []ArticleInput) ([]ClassificationResult, error)
	GenerateDaySummary(ctx context.Context, category string, articles []ArticleInput) (DayBriefing, error)
}
