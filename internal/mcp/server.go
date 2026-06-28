package mcp

import (
	"context"
	"fmt"
	"net/http"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type McpServer struct {
	Handler *gomcp.StreamableHTTPHandler
	srv     *gomcp.Server
}

func NewMcpServer(version string) *McpServer {
	server := gomcp.NewServer(&gomcp.Implementation{Name: "traggo-mcp-integration", Version: version}, nil)

	server.AddTool(&gomcp.Tool{
		Name:        "ping_auth",
		Description: "A simple ping tool that returns the provided auth token to verify middleware injection",
		InputSchema: map[string]any{
			"type": "object",
		},
	}, func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		token, ok := ctx.Value("traggo_token").(string)
		if !ok || token == "" {
			return nil, fmt.Errorf("unauthorized: missing token in context")
		}

		responseText := fmt.Sprintf("Success! Tool executed with token: %s", token)
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: responseText}},
		}, nil
	})

	return &McpServer{
		Handler: gomcp.NewStreamableHTTPHandler(func(r *http.Request) *gomcp.Server {
			return server
		}, nil),
		srv: server,
	}

}
