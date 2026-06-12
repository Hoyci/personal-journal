package core

import "sort"

func SelectTopArticles(sections []Section, total int) []Section {
	if len(sections) == 0 {
		return sections
	}

	base := total / len(sections)
	extra := total % len(sections)

	// Ordenar seções por score médio para dar os artigos extras às mais relevantes
	type sectionScore struct {
		index int
		avg   float64
	}
	scores := make([]sectionScore, len(sections))
	for i, s := range sections {
		scores[i] = sectionScore{index: i, avg: avgScore(s.Articles)}
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].avg > scores[j].avg
	})

	// Distribuir cotas
	quotas := make([]int, len(sections))
	for i := range quotas {
		quotas[i] = base
	}
	for i := 0; i < extra; i++ {
		quotas[scores[i].index]++
	}

	// Aplicar cotas — artigos já vêm ordenados por score do classifier
	result := make([]Section, len(sections))
	for i, s := range sections {
		quota := quotas[i]
		if quota > len(s.Articles) {
			quota = len(s.Articles)
		}
		result[i] = Section{
			Category:   s.Category,
			Articles:   s.Articles[:quota],
			Summary:    s.Summary,
			Highlights: s.Highlights,
		}
	}

	return result
}

func avgScore(articles []Article) float64 {
	if len(articles) == 0 {
		return 0
	}
	var sum float64
	for _, a := range articles {
		sum += a.RelevanceScore
	}
	return sum / float64(len(articles))
}
