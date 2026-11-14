package aibot

import "sync"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type SessionMemory struct {
	mu       sync.RWMutex
	sessions map[string][]Message
	maxTurns int
}

func NewSessionMemory(maxTurns int) *SessionMemory {
	if maxTurns <= 0 {
		maxTurns = 10
	}
	return &SessionMemory{
		sessions: make(map[string][]Message),
		maxTurns: maxTurns,
	}
}

func (m *SessionMemory) Append(sessionID string, msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := append(m.sessions[sessionID], msg)
	if len(history) > m.maxTurns {
		history = history[len(history)-m.maxTurns:]
	}
	m.sessions[sessionID] = history
}

func (m *SessionMemory) GetHistory(sessionID string) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	src := m.sessions[sessionID]
	copied := make([]Message, len(src))
	copy(copied, src)
	return copied
}

func (m *SessionMemory) Clear(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}
