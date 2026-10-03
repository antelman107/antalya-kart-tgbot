package advisor

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLiveGeminiKey(t *testing.T) {
	if os.Getenv("ADK_LIVE") == "" {
		t.Skip("set ADK_LIVE=1 to call Gemini")
	}
	key := os.Getenv("GOOGLE_API_KEY")
	if key == "" {
		t.Fatal("GOOGLE_API_KEY is required")
	}
	mcpURL := os.Getenv("ANTALYAKART_MCP_URL")
	if mcpURL == "" {
		mcpURL = DefaultMCPURL
	}
	agent, err := NewAgent(t.Context(), key, os.Getenv("GEMINI_MODEL"), mcpURL)
	if err != nil {
		t.Fatal(redactLive(err, key))
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	reply, err := agent.Reply(ctx, "live", nil, "Reply with the single word pong")
	if err != nil {
		t.Fatal(redactLive(err, key))
	}
	if !strings.Contains(strings.ToLower(reply), "pong") {
		t.Fatalf("reply = %q", reply)
	}
}

func redactLive(err error, key string) string {
	return strings.ReplaceAll(err.Error(), key, "[REDACTED]")
}
