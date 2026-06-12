package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	anthropicAPI = "https://api.anthropic.com/v1/messages"
	model        = "claude-haiku-4-5"
)

type AnthropicProvider struct {
	apiKey string
	client *http.Client
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
		client: &http.Client{},
	}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (a *AnthropicProvider) call(ctx context.Context, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: 4000,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPI, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API retornou status %d", resp.StatusCode)
	}

	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return result.Content[0].Text, nil
}

func (a *AnthropicProvider) ClassifyArticles(ctx context.Context, category string, articles []ArticleInput) ([]ClassificationResult, error) {
	const batchSize = 50
	var allResults []ClassificationResult

	for i := 0; i < len(articles); i += batchSize {
		end := i + batchSize
		if end > len(articles) {
			end = len(articles)
		}
		batch := articles[i:end]

		prompt := buildClassifyPrompt(category, batch)
		text, err := a.call(ctx, prompt)
		if err != nil {
			return nil, err
		}

		results, err := parseClassifyResponse(text, batch)
		if err != nil {
			return nil, err
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func (a *AnthropicProvider) GenerateDaySummary(ctx context.Context, category string, articles []ArticleInput) (DayBriefing, error) {
	prompt := buildSummaryPrompt(category, articles)
	text, err := a.call(ctx, prompt)
	if err != nil {
		return DayBriefing{}, err
	}
	return parseSummaryResponse(text)
}

func buildClassifyPrompt(category string, articles []ArticleInput) string {
	list := ""
	for _, a := range articles {
		list += fmt.Sprintf("[%d] %s\nFonte: %s\n%s\n\n", a.Index, a.Title, a.Source, a.Summary)
	}

	return fmt.Sprintf(`Você é um assistente que classifica artigos de notícias.

Categoria: %s

Artigos:
%s

Para cada artigo, avalie:
- Se é relevante (não é ruído como celebridades, esportes, entretenimento fútil)
- Score de relevância de 0.0 a 10.0

Responda APENAS com JSON válido, sem texto antes ou depois:
{
  "results": [
    {"index": 0, "relevant": true, "score": 8.5, "reasoning": "motivo breve"},
    {"index": 1, "relevant": false, "score": 1.0, "reasoning": "motivo breve"}
  ]
}`, category, list)
}

func buildSummaryPrompt(category string, articles []ArticleInput) string {
	list := ""
	for _, a := range articles {
		list += fmt.Sprintf("- %s\n", a.Title)
	}

	return fmt.Sprintf(`Você é um assistente que gera resumos matinais de notícias.

Categoria: %s

Artigos do dia:
%s

Gere um resumo executivo do que aconteceu hoje nessa categoria.

Formate o summary e os highlights usando HTML suportado pelo Telegram.
Regras:
- Use <b>negrito</b> para termos importantes.
- Use <i>itálico</i> para ênfase secundária.
- NÃO tente escapar caracteres (como barras invertidas). Apenas use texto puro e as tags HTML permitidas.

Responda APENAS com JSON válido, sem texto antes ou depois:
{
  "summary": "parágrafo de 3-4 linhas resumindo o dia",
  "highlights": ["destaque 1", "destaque 2", "destaque 3"]
}`, category, list)
}

func parseClassifyResponse(text string, articles []ArticleInput) ([]ClassificationResult, error) {
	clean := stripMarkdownJSON(text)
	var response struct {
		Results []ClassificationResult `json:"results"`
	}
	if err := json.Unmarshal([]byte(clean), &response); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta da IA: %w\nResposta: %s", err, text)
	}
	return response.Results, nil
}

func parseSummaryResponse(text string) (DayBriefing, error) {
	clean := stripMarkdownJSON(text)
	var briefing DayBriefing
	if err := json.Unmarshal([]byte(clean), &briefing); err != nil {
		return DayBriefing{}, fmt.Errorf("erro ao parsear resumo da IA: %w\nResposta: %s", err, text)
	}
	return briefing, nil
}

func stripMarkdownJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		first := strings.Index(text, "\n")
		if first != -1 {
			text = text[first+1:]
		}
		if last := strings.LastIndex(text, "```"); last != -1 {
			text = text[:last]
		}
	}
	return strings.TrimSpace(text)
}
