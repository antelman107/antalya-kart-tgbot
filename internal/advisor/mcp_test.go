package advisor

import (
	"context"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"

	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool/mcptoolset"
)

func TestMCPToolsetListsServerTools(t *testing.T) {
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "antalyakart", Version: "test"}, nil)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "search_routes_and_stops",
		Description: "Search routes and stops.",
	}, func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "ok"}},
		}, nil, nil
	})
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(func() {
		// The streamable client keeps a GET connection open. Close it first,
		// otherwise Server.Close waits on that connection until the test times out.
		httpServer.CloseClientConnections()
		httpServer.Close()
	})

	toolset, err := mcptoolset.New(mcptoolset.Config{Endpoint: httpServer.URL})
	if err != nil {
		t.Fatal(err)
	}
	tools, err := toolset.Tools(readonlyCtx{Context: t.Context()})
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name() != "search_routes_and_stops" {
		t.Fatalf("tools = %#v", tools)
	}
}

type readonlyCtx struct{ context.Context }

func (readonlyCtx) UserContent() *genai.Content { return nil }
func (readonlyCtx) InvocationID() string        { return "test" }
func (readonlyCtx) AgentName() string           { return "test" }
func (readonlyCtx) ReadonlyState() session.ReadonlyState {
	return emptyState{}
}
func (readonlyCtx) UserID() string    { return "user" }
func (readonlyCtx) AppName() string   { return "app" }
func (readonlyCtx) SessionID() string { return "session" }
func (readonlyCtx) Branch() string    { return "" }

type emptyState struct{}

func (emptyState) Get(string) (any, error) { return nil, session.ErrStateKeyNotExist }
func (emptyState) All() iter.Seq2[string, any] {
	return func(func(string, any) bool) {}
}
