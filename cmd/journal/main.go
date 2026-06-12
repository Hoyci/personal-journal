package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hoyci/personal-journal/internal/ai"
	"github.com/hoyci/personal-journal/internal/collector"
	"github.com/hoyci/personal-journal/internal/config"
	"github.com/hoyci/personal-journal/internal/core"
	"github.com/hoyci/personal-journal/internal/notifier"
	"github.com/hoyci/personal-journal/internal/pipeline"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	godotenv.Load()

	root := &cobra.Command{
		Use:   "journal",
		Short: "Seu jornal pessoal com IA",
	}

	root.AddCommand(sendCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func sendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send",
		Short: "Coleta, gera e envia o briefing do dia via Telegram",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// 1. Inicializa o motor
			engine, err := setupPipeline()
			if err != nil {
				return err
			}

			// 2. Roda o fluxo principal
			briefing, err := engine.Run(ctx)
			if err != nil {
				return err
			}

			// 3. Envia a notificação
			if err := sendNotification(ctx, briefing); err != nil {
				return err
			}

			return nil
		},
	}
}

// setupPipeline cuida apenas de instanciar e conectar as peças do sistema
func setupPipeline() (*pipeline.Pipeline, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar config: %w", err)
	}

	var collectors []collector.Collector
	for _, s := range cfg.Sources {
		cat := core.Category{
			Name:     s.Category,
			Priority: s.Priority,
		}
		collectors = append(collectors, collector.NewRSSCollector(s.Name, s.URL, cat))
	}

	llm := ai.NewAnthropicProvider()

	return pipeline.New(collectors, llm), nil
}

// sendNotification abstrai o envio da mensagem formatada
func sendNotification(ctx context.Context, briefing core.Briefing) error {
	n, err := notifier.NewTelegramNotifier(ctx)
	if err != nil {
		return err
	}

	message := notifier.FormatBriefing(briefing)

	if err := n.SendToMe(ctx, message); err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}

	log.Printf("INFO: briefing enviado via Telegram com sucesso!")
	return nil
}
