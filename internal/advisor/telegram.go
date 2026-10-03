package advisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var telegramAPI = "https://api.telegram.org"

// TelegramClient talks to the Bot API. The token stays in memory and is not logged.
type TelegramClient struct {
	http   *http.Client
	token  string
	secret string
}

func NewTelegramClient(token, webhookSecret string) *TelegramClient {
	return &TelegramClient{
		http:   &http.Client{Timeout: 20 * time.Second},
		token:  strings.TrimSpace(token),
		secret: webhookSecret,
	}
}

func (c *TelegramClient) SetWebhook(ctx context.Context, publicURL string) error {
	body := map[string]any{
		"url":             publicURL,
		"allowed_updates": []string{"message", "callback_query"},
	}
	if c.secret != "" {
		body["secret_token"] = c.secret
	}
	return c.call(ctx, "setWebhook", body, nil)
}

func (c *TelegramClient) SendText(ctx context.Context, chatID int64, replyTo int, text string) error {
	for _, chunk := range splitTelegramText(text) {
		payload := map[string]any{
			"chat_id":    chatID,
			"text":       chunk,
			"parse_mode": "HTML",
		}
		if replyTo != 0 {
			payload["reply_parameters"] = map[string]any{"message_id": replyTo}
			replyTo = 0
		}
		if err := c.call(ctx, "sendMessage", payload, nil); err != nil {
			if !isHTMLParseError(err) {
				return err
			}
			delete(payload, "parse_mode")
			if err := c.call(ctx, "sendMessage", payload, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *TelegramClient) NotifyTyping(ctx context.Context, chatID int64) error {
	return c.call(ctx, "sendChatAction", map[string]any{
		"chat_id": chatID,
		"action":  "typing",
	}, nil)
}

func (c *TelegramClient) AnswerCallback(ctx context.Context, callbackID string) error {
	if callbackID == "" {
		return nil
	}
	return c.call(ctx, "answerCallbackQuery", map[string]any{
		"callback_query_id": callbackID,
	}, nil)
}

func (c *TelegramClient) call(ctx context.Context, method string, payload any, out any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.methodURL(method), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var envelope struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("telegram %s failed: %s", method, resp.Status)
		}
		return fmt.Errorf("telegram %s returned invalid json", method)
	}
	if resp.StatusCode != http.StatusOK || !envelope.OK {
		if envelope.Description == "" {
			envelope.Description = resp.Status
		}
		return fmt.Errorf("telegram %s rejected the call: %s", method, envelope.Description)
	}
	if out != nil && len(envelope.Result) > 0 {
		return json.Unmarshal(envelope.Result, out)
	}
	return nil
}

func (c *TelegramClient) methodURL(method string) string {
	return telegramAPI + "/bot" + c.token + "/" + method
}

var (
	telegramTagPattern = regexp.MustCompile(`</?(?:b|strong|i|em|u|ins|s|strike|del|code|pre|a|blockquote|tg-spoiler)\b`)
	markdownBold       = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	markdownCode       = regexp.MustCompile("`([^`\n]+)`")
	markdownList       = regexp.MustCompile(`(?m)^[ \t]*[*\-][ \t]+`)
)

// telegramHTML turns a model reply into Telegram HTML.
// A reply that already uses the allowed tags is kept. Markdown is converted,
// and plain text is escaped so parse_mode HTML does not reject it.
func telegramHTML(text string) string {
	text = strings.TrimSpace(text)
	if text == "" || telegramTagPattern.MatchString(text) {
		return text
	}
	escaped := html.EscapeString(text)
	escaped = markdownList.ReplaceAllString(escaped, "• ")
	escaped = markdownBold.ReplaceAllString(escaped, "<b>$1</b>")
	escaped = markdownCode.ReplaceAllString(escaped, "<code>$1</code>")
	return escaped
}

func isHTMLParseError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "can't parse entities") || strings.Contains(msg, "cant parse entities")
}

func splitTelegramText(text string) []string {
	const limit = 4000
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"..."}
	}
	if len(text) <= limit {
		return []string{text}
	}
	var chunks []string
	for len(text) > limit {
		chunks = append(chunks, text[:limit])
		text = text[limit:]
	}
	if text != "" {
		chunks = append(chunks, text)
	}
	return chunks
}
