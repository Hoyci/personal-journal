package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"

	"github.com/hoyci/personal-journal/internal/core"
)

const telegramAPIBase = "https://api.telegram.org/bot%s/sendMessage"

type TelegramNotifier struct {
	token  string
	client *http.Client
}

func NewTelegramNotifier(_ context.Context) (*TelegramNotifier, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN não definido no .env")
	}

	return &TelegramNotifier{
		token:  token,
		client: &http.Client{},
	}, nil
}

func (t *TelegramNotifier) SendToMe(ctx context.Context, message string) error {
	chatID := os.Getenv("TELEGRAM_MY_CHAT_ID")
	if chatID == "" {
		return fmt.Errorf("TELEGRAM_MY_CHAT_ID não definido no .env")
	}
	return t.SendMessage(ctx, chatID, message)
}

func (t *TelegramNotifier) SendMessage(ctx context.Context, chatID string, message string) error {
	chunks := splitMessage(message, 4096)
	for _, chunk := range chunks {
		if err := t.sendChunk(ctx, chatID, chunk); err != nil {
			return err
		}
	}
	return nil
}

func (t *TelegramNotifier) sendChunk(ctx context.Context, chatID, text string) error {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	url := fmt.Sprintf(telegramAPIBase, t.token)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("telegram API retornou status %d: %v", resp.StatusCode, result)
	}

	return nil
}

func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		cut := maxLen
		if idx := strings.LastIndex(text[:cut], "\n"); idx > 0 {
			cut = idx
		}

		chunks = append(chunks, text[:cut])
		text = strings.TrimSpace(text[cut:])
	}
	return chunks
}

func FormatBriefing(briefing core.Briefing) string {
	var sb strings.Builder

	// Título principal
	sb.WriteString(fmt.Sprintf("<b>BRIEFING — %s</b>\n", briefing.Date.Format("02/01/2006")))

	for _, section := range briefing.Sections {
		// Nome da Categoria protegido
		sb.WriteString(fmt.Sprintf("\n<b>━━ %s ━━</b>\n", html.EscapeString(strings.ToUpper(section.Category.Name))))

		if section.Summary != "" {
			// A IA já manda com as tags <b> e <i>, então imprimimos direto
			sb.WriteString(fmt.Sprintf("💡 %s\n", section.Summary))
		}

		if len(section.Highlights) > 0 {
			sb.WriteString("\n<b>Destaques:</b>\n")
			for _, h := range section.Highlights {
				sb.WriteString(fmt.Sprintf("• %s\n", h))
			}
		}

		sb.WriteString("\n")
		for _, a := range section.Articles {
			// Título e Fonte protegidos com html.EscapeString
			sb.WriteString(fmt.Sprintf(
				"<b>%.0f/10</b> — <a href=\"%s\">%s</a>\n📰 %s | %s\n\n",
				a.RelevanceScore,
				a.URL,
				html.EscapeString(a.Title),
				html.EscapeString(a.Source),
				a.PublishedAt.Format("02/01 15:04"),
			))
		}
	}

	return strings.TrimSpace(sb.String())
}
