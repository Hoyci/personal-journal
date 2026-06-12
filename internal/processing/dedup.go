package processing

import "github.com/hoyci/personal-journal/internal/core"

type Deduplicator struct {
	seen map[string]bool
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{seen: make(map[string]bool)}
}

func (d *Deduplicator) Deduplicate(articles []core.Article) []core.Article {
	var unique []core.Article

	for _, a := range articles {
		if d.seen[a.URL] {
			continue
		}
		d.seen[a.URL] = true
		unique = append(unique, a)
	}

	return unique
}
