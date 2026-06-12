package pipeline

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/hoyci/personal-journal/internal/ai"
	"github.com/hoyci/personal-journal/internal/collector"
	"github.com/hoyci/personal-journal/internal/core"
	"github.com/hoyci/personal-journal/internal/processing"
)

// Pipeline orquestra o fluxo de coleta, processamento e geração do briefing.
type Pipeline struct {
	collectors []collector.Collector
	llm        ai.LLMProvider
}

func New(collectors []collector.Collector, llm ai.LLMProvider) *Pipeline {
	return &Pipeline{
		collectors: collectors,
		llm:        llm,
	}
}

// Run executa todo o fluxo de ponta a ponta.
func (p *Pipeline) Run(ctx context.Context) (core.Briefing, error) {
	articles := p.fetchConcurrent(ctx)
	articles = p.processArticles(articles)

	grouped := p.groupByCategory(articles)
	sections := p.processSectionsWithAI(ctx, grouped)
	sections = core.SelectTopArticles(sections, 10)

	log.Printf("INFO: briefing gerado com sucesso")

	return core.Briefing{
		Date:     time.Now(),
		Sections: sections,
	}, nil
}

func (p *Pipeline) fetchConcurrent(ctx context.Context) []core.Article {
	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	type result struct {
		articles []core.Article
		name     string
		err      error
	}

	results := make(chan result, len(p.collectors))
	for _, c := range p.collectors {
		go func(col collector.Collector) {
			articles, err := col.Fetch(fetchCtx)
			results <- result{articles: articles, name: col.Name(), err: err}
		}(c)
	}

	var allArticles []core.Article
	for range p.collectors {
		r := <-results
		if r.err != nil {
			log.Printf("WARN: falha ao coletar %s: %v", r.name, r.err)
			continue
		}
		log.Printf("INFO: coletado %s — %d artigos", r.name, len(r.articles))
		allArticles = append(allArticles, r.articles...)
	}

	return allArticles
}

func (p *Pipeline) processArticles(articles []core.Article) []core.Article {
	dedup := processing.NewDeduplicator()
	uniqueArticles := dedup.Deduplicate(articles)
	log.Printf("INFO: %d artigos após deduplicação", len(uniqueArticles))

	now := time.Now()
	todayLocal := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	tomorrowLocal := todayLocal.Add(24 * time.Hour)

	var filtered []core.Article
	for _, a := range uniqueArticles {
		pub := a.PublishedAt.In(time.Local)
		if !pub.Before(todayLocal) && pub.Before(tomorrowLocal) {
			filtered = append(filtered, a)
		}
	}

	log.Printf("INFO: %d artigos de hoje após filtro de data", len(filtered))
	return filtered
}

func (p *Pipeline) processSectionsWithAI(ctx context.Context, sections []core.Section) []core.Section {
	aiCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	type aiResult struct {
		section core.Section
		err     error
	}

	results := make(chan aiResult, len(sections))

	for _, section := range sections {
		go func(s core.Section) {
			log.Printf("INFO: classificando categoria %s (%d artigos)...", s.Category.Name, len(s.Articles))

			inputs := p.toArticleInputs(s.Articles)
			classifications, err := p.llm.ClassifyArticles(aiCtx, s.Category.Name, inputs)
			if err != nil {
				log.Printf("WARN: falha ao classificar %s: %v", s.Category.Name, err)
				results <- aiResult{section: s}
				return
			}

			relevantArticles := p.applyClassifications(s.Articles, classifications)
			if len(relevantArticles) == 0 {
				results <- aiResult{}
				return
			}

			relevantInputs := p.toArticleInputs(relevantArticles)
			dayBriefing, err := p.llm.GenerateDaySummary(aiCtx, s.Category.Name, relevantInputs)
			if err != nil {
				log.Printf("WARN: falha ao gerar resumo de %s: %v", s.Category.Name, err)
			}

			log.Printf("INFO: %s — %d relevantes de %d", s.Category.Name, len(relevantArticles), len(s.Articles))

			results <- aiResult{section: core.Section{
				Category:   s.Category,
				Articles:   relevantArticles,
				Summary:    dayBriefing.Summary,
				Highlights: dayBriefing.Highlights,
			}}
		}(section)
	}

	var processedSections []core.Section
	for range sections {
		r := <-results
		if len(r.section.Articles) > 0 {
			processedSections = append(processedSections, r.section)
		}
	}

	return processedSections
}

// Helpers que antes poluíam o main.go
func (p *Pipeline) toArticleInputs(articles []core.Article) []ai.ArticleInput {
	inputs := make([]ai.ArticleInput, len(articles))
	for i, a := range articles {
		inputs[i] = ai.ArticleInput{
			Index:   i,
			Title:   a.Title,
			Summary: a.Content,
			Source:  a.Source,
		}
	}
	return inputs
}

func (p *Pipeline) applyClassifications(articles []core.Article, classifications []ai.ClassificationResult) []core.Article {
	scoreMap := make(map[int]ai.ClassificationResult)
	for _, c := range classifications {
		scoreMap[c.Index] = c
	}

	var relevant []core.Article
	for i, a := range articles {
		c, ok := scoreMap[i]
		if !ok || !c.Relevant {
			continue
		}
		a.RelevanceScore = c.Score
		relevant = append(relevant, a)
	}

	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].RelevanceScore > relevant[j].RelevanceScore
	})

	return relevant
}

func (p *Pipeline) groupByCategory(articles []core.Article) []core.Section {
	var order []string
	byCategory := make(map[string]*core.Section)

	for _, a := range articles {
		name := a.Category.Name
		if _, exists := byCategory[name]; !exists {
			byCategory[name] = &core.Section{Category: a.Category}
			order = append(order, name)
		}
		byCategory[name].Articles = append(byCategory[name].Articles, a)
	}

	var sections []core.Section
	for _, name := range order {
		sections = append(sections, *byCategory[name])
	}
	return sections
}
