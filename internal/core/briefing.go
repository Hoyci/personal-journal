package core

import "time"

type Briefing struct {
	Date     time.Time
	Sections []Section
}

type Section struct {
	Category   Category
	Articles   []Article
	Summary    string
	Highlights []string
}
