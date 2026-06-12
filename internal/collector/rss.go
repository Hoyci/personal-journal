package collector

import (
	"context"
	"strings"
	"time"

	"github.com/hoyci/personal-journal/internal/core"
	"github.com/hoyci/personal-journal/internal/processing"
	"github.com/mmcdole/gofeed"
)

type RSSCollector struct {
	name     string
	url      string
	category core.Category
}

func NewRSSCollector(name, url string, category core.Category) *RSSCollector {
	return &RSSCollector{
		name:     name,
		url:      url,
		category: category,
	}
}

func (r *RSSCollector) Name() string {
	return r.name
}

func (r *RSSCollector) Fetch(ctx context.Context) ([]core.Article, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(r.url, ctx)
	if err != nil {
		return nil, err
	}

	var articles []core.Article
	for _, item := range feed.Items {
		a := core.Article{
			URL:         item.Link,
			Title:       item.Title,
			Content:     processing.StripHTML(item.Description),
			Source:      r.name,
			Category:    r.category,
			CollectedAt: time.Now(),
		}

		if item.GUID != "" && strings.HasPrefix(item.GUID, "http") {
			a.URL = item.GUID
		}

		if item.PublishedParsed != nil {
			a.PublishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			a.PublishedAt = *item.UpdatedParsed
		} else {
			a.PublishedAt = time.Now()
		}

		articles = append(articles, a)
	}

	return articles, nil
}
