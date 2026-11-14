package aibot

import (
	"context"
	"testing"
)

type mockAIProvider struct{}

func (m *mockAIProvider) GenerateReply(_ context.Context, history []Message) (string, error) {
	last := history[len(history)-1].Content
	return "Mock AI Reply para: " + last, nil
}

func TestBotService_ConversationContext(t *testing.T) {
	mem := NewSessionMemory(10)
	svc := NewBotService(&mockAIProvider{}, mem)
	ctx := context.Background()

	reply1, hist1, err := svc.Chat(ctx, "user-123", "Olá, quero falar sobre Golang")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if reply1 == "" || len(hist1) != 2 {
		t.Fatalf("esperava 2 mensagens no histórico (user + assistant), obteve %d", len(hist1))
	}

	_, hist2, err := svc.Chat(ctx, "user-123", "Como funcionam as goroutines?")
	if err != nil {
		t.Fatalf("erro inesperado no segundo turno: %v", err)
	}
	if len(hist2) != 4 {
		t.Errorf("esperava 4 mensagens acumuladas na sessão, obteve %d", len(hist2))
	}
}
