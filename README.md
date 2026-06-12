# 📰 Jornal Pessoal com IA

Um sistema automatizado e *serverless* construído em Go que coleta notícias de múltiplas fontes, filtra o ruído usando Inteligência Artificial e entrega um briefing diário executivo diretamente no seu Telegram.

A ideia central é simples: você acorda, e o seu jornal personalizado, limpo e direto ao ponto já está no seu celular, pronto para ser lido em menos de 15 minutos.

---

## ✨ Funcionalidades

* **Coleta Multi-fonte:** Lê feeds RSS de tecnologia, política e backend (ou qualquer outra área que você configurar).
* **Filtro Inteligente:** Utiliza a API da Anthropic (Claude) para pontuar a relevância de cada artigo de acordo com seus interesses, ignorando ruídos (como esportes ou fofocas).
* **Resumos Executivos:** A IA lê os artigos mais relevantes e gera um resumo rápido e direto para cada categoria.
* **Deduplicação:** Evita que a mesma notícia, vinda de fontes diferentes, polua sua leitura.
* **Entrega via Telegram:** Formatação limpa em HTML diretamente em uma conversa com o seu bot do Telegram.
* **Piloto Automático (Serverless):** Roda gratuitamente e diariamente via GitHub Actions, sem precisar de um servidor ligado 24/7.

---

## 🏗️ Arquitetura e Fluxo

1. **Scheduler (GitHub Actions):** Dispara a pipeline automaticamente no horário configurado (ex: 12h00).
2. **Collector:** Busca os artigos do dia nas fontes definidas no `config.yaml`.
3. **Processor:** Limpa as tags HTML residuais, unifica o formato e remove artigos duplicados ou de dias anteriores.
4. **AI Classifier (Anthropic):** Lê os títulos e trechos, dando uma nota de 0 a 10 e separando apenas o que importa.
5. **AI Summarizer (Anthropic):** Gera um parágrafo de resumo geral e destaca os pontos principais da categoria.
6. **Notifier (Telegram):** Formata os dados e envia as mensagens fragmentadas e validadas para o seu celular.

---

## 🚀 Como rodar localmente

### Pré-requisitos

* [Go](https://go.dev/) 1.25+ instalado.
* Uma chave de API da [Anthropic](https://console.anthropic.com/).
* Um Bot no Telegram (criado via [BotFather](https://t.me/botfather)) e o seu Chat ID.

### Passo 1: Configuração das Credenciais

Crie um arquivo chamado `.env` na raiz do projeto com as seguintes variáveis:

```env
ANTHROPIC_API_KEY=sua-chave-api-aqui
TELEGRAM_BOT_TOKEN=seu-token-do-botfather-aqui
TELEGRAM_MY_CHAT_ID=seu-id-do-telegram-aqui
```

> **Nota:** O arquivo `.env` já está no `.gitignore` para garantir sua segurança.

### Passo 2: Configuração das Fontes

O arquivo `config.yaml` na raiz do projeto define o que o bot deve ler. Você pode adicionar ou remover fontes à vontade:

```yaml
sources:
  - name: "Hacker News"
    url: "https://hnrss.org/frontpage"
    category: "Tech"
    priority: "high"
  # Adicione mais fontes conforme sua necessidade...
```

### Passo 3: Executando

Para rodar a pipeline inteira e receber o briefing no Telegram:

```bash
go run ./cmd/journal/main.go send
```

---

## ☁️ Automação com GitHub Actions

O projeto está configurado para rodar na nuvem, de forma totalmente gratuita, utilizando o GitHub Actions.

**Para ativar:**

1. Faça o push do seu código para o GitHub (garanta que o arquivo `.github/workflows/daily-journal.yml` esteja no repositório).
2. Vá em **Settings > Secrets and variables > Actions**.
3. Adicione as suas três credenciais como **Repository Secrets**:
   * `ANTHROPIC_API_KEY`
   * `TELEGRAM_BOT_TOKEN`
   * `TELEGRAM_MY_CHAT_ID`

A action rodará automaticamente conforme o agendamento no arquivo `.yml`, ou pode ser acionada manualmente na aba **Actions**.
