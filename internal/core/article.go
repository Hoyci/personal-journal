package core

import "time"

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
	PriorityIgnore Priority = "ignore"
)

type Category struct {
	Name     string
	Priority Priority
}

type Article struct {
	URL            string
	Title          string
	Content        string
	Summary        string
	Source         string
	Category       Category
	RelevanceScore float64
	PublishedAt    time.Time
	CollectedAt    time.Time
}
