package advisor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelegramHTMLConvertsMarkdownAndKeepsTags(t *testing.T) {
	got := telegramHTML("Рядом **Boğa & Co**\n* **BOĞAÇAY CD-5** (ID: `12347`)")
	want := "Рядом <b>Boğa &amp; Co</b>\n• <b>BOĞAÇAY CD-5</b> (ID: <code>12347</code>)"
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}

	already := "Остановка <b>Otogar</b>, id <code>10828</code>"
	if telegramHTML(already) != already {
		t.Fatalf("html reply was rewritten: %q", telegramHTML(already))
	}

	if got := telegramHTML("a < b & c"); got != "a &lt; b &amp; c" {
		t.Fatalf("plain text = %q", got)
	}
}

func TestSendTextUsesHTMLAndFallsBackWhenTelegramRejectsIt(t *testing.T) {
	var modes []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("json: %v", err)
		}
		mode, _ := payload["parse_mode"].(string)
		modes = append(modes, mode)
		if mode == "HTML" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"description":"Bad Request: can't parse entities"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	previous := telegramAPI
	telegramAPI = server.URL
	t.Cleanup(func() { telegramAPI = previous })

	client := NewTelegramClient("token", "")
	if err := client.SendText(context.Background(), 7, 3, "<b>broken"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(modes, ",") != "HTML," {
		t.Fatalf("parse modes = %#v", modes)
	}
}
