package aibot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AIProvider interface {
	GenerateReply(ctx context.Context, history []Message) (string, error)
}

type GeminiProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetries int
}

func NewGeminiProvider(apiKey, model string, client *http.Client) *GeminiProvider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &GeminiProvider{
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
		maxRetries: 2,
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

func (p *GeminiProvider) GenerateReply(ctx context.Context, history []Message) (string, error) {
	if len(history) == 0 {
		return "", errors.New("histórico de mensagens vazio")
	}

	// Modo Fallback inteligente caso GEMINI_API_KEY não esteja definida (facilita testes de recrutadores)
	if strings.TrimSpace(p.apiKey) == "" {
		last := history[len(history)-1].Content
		return fmt.Sprintf("[Modo Demo Local - Sem API Key] Recebi a sua mensagem (%q). Contexto ativo com %d mensagens na sessão.", last, len(history)), nil
	}

	var contents []geminiContent
	for _, m := range history {
		role := "user"
		if m.Role == "assistant" || m.Role == "model" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	payload, err := json.Marshal(geminiRequest{Contents: contents})
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(attempt*200) * time.Millisecond):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("erro temporário da API externa (status %d)", resp.StatusCode)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("falha na API de IA: status %d", resp.StatusCode)
		}

		var parsed geminiResponse
		err = json.NewDecoder(resp.Body).Decode(&parsed)
		resp.Body.Close()
		if err != nil {
			return "", err
		}

		if len(parsed.Candidates) > 0 && len(parsed.Candidates[0].Content.Parts) > 0 {
			return parsed.Candidates[0].Content.Parts[0].Text, nil
		}
		return "", errors.New("resposta vazia do modelo de IA")
	}

	return "", lastErr
}
