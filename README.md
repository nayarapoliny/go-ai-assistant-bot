# 🤖 Go AI Assistant Bot Microservice

Microsserviço e CLI de Assistente de Inteligência Artificial construído em **Go (Golang)** com integração REST/JSON à API **Google Gemini**, gestão de histórico conversacional *Thread-Safe* por sessão e mecanismo de *Retry* resiliente.

## 🚀 Destaques Técnicos

- **Gestão de Estado Conversacional Concorrente:** Armazenamento de sessões em memória protegido por `sync.RWMutex` com janela deslizante de contexto (`maxTurns`).
- **Integração com LLM & Resiliência:** Cliente HTTP preparado para consumir a API do Gemini com *Exponential Backoff Retry* para falhas transitórias (`5xx` / `429`) e modo de *Fallback* automático para testes locais sem chave de API.
- **Arquitetura Desacoplada:** Interface `AIProvider` permitindo substituir o motor de IA ou injetar *mocks* em testes unitários.
- **Modo Híbrido (REST API + CLI Interativo):** Pode ser executado como microsserviço HTTP (`POST /api/v1/chat`) ou diretamente no terminal (`-cli`).

## 🛠️ Como Executar e Testar

`go test -v -race ./...`

### 1. Modo Interativo no Terminal (CLI)
`go run ./cmd/bot -cli`

### 2. Modo Microsserviço HTTP
`GEMINI_API_KEY="sua_chave_opcional" go run ./cmd/bot -port 8081`

Exemplo de requisição:
`curl -X POST http://localhost:8081/api/v1/chat -H "Content-Type: application/json" -d '{"session_id":"user-1","prompt":"Explique como funcionam os channels em Go"}'`
