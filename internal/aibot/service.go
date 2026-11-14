package aibot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type BotService struct {
	provider AIProvider
	memory   *SessionMemory
}

func NewBotService(provider AIProvider, memory *SessionMemory) *BotService {
	return &BotService{
		provider: provider,
		memory:   memory,
	}
}

func (s *BotService) Chat(ctx context.Context, sessionID, userPrompt string) (string, []Message, error) {
	userPrompt = strings.TrimSpace(userPrompt)
	if userPrompt == "" {
		return "", nil, errors.New("o campo 'prompt' não pode estar vazio")
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "default-session"
	}

	s.memory.Append(sessionID, Message{Role: "user", Content: userPrompt})
	history := s.memory.GetHistory(sessionID)

	reply, err := s.provider.GenerateReply(ctx, history)
	if err != nil {
		return "", nil, err
	}

	s.memory.Append(sessionID, Message{Role: "assistant", Content: reply})
	return reply, s.memory.GetHistory(sessionID), nil
}

type chatRequest struct {
	SessionID string `json:"session_id"`
	Prompt    string `json:"prompt"`
}

type chatResponse struct {
	SessionID   string    `json:"session_id"`
	Reply       string    `json:"reply"`
	HistorySize int       `json:"history_size"`
	History     []Message `json:"history"`
}

func (s *BotService) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
			return
		}

		if req.SessionID == "" {
			req.SessionID = "default-session"
		}

		reply, history, err := s.Chat(r.Context(), req.SessionID, req.Prompt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, chatResponse{
			SessionID:   req.SessionID,
			Reply:       reply,
			HistorySize: len(history),
			History:     history,
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
