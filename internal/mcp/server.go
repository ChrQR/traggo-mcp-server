package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ChrQR/traggo-mcp-server/internal/traggo"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type McpServer struct {
	Handler *gomcp.StreamableHTTPHandler
	srv     *gomcp.Server
}

// jsonResult renders v as pretty JSON inside a tool text result.
func jsonResult(v any) (*gomcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func NewMcpServer(version, traggoURL string) *McpServer {
	server := gomcp.NewServer(&gomcp.Implementation{Name: "traggo-mcp-integration", Version: version}, nil)

	svc := traggo.NewService(traggoURL)

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

	gomcp.AddTool(server, &gomcp.Tool{
		Name: "add_timespan",
		Description: `Record a completed timespan with a start, end, tags and a note.
		Omit end to leave the timespan running. Tags are a list of key/value pairs.`,
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.AddTimespanInput) (*gomcp.CallToolResult, any, error) {
		ts, err := svc.AddTimespan(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(ts)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name: "start_timer",
		Description: `Start a running timer (an open timespan with no end) with optional tags and a note.
		Start defaults to now if omitted. Use stop_timer to close it.`,
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.StartTimerInput) (*gomcp.CallToolResult, any, error) {
		ts, err := svc.StartTimer(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(ts)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name: "stop_timer",
		Description: `Stop a running timer by its id (from list_timers).
		End defaults to now if omitted.`,
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.StopTimerInput) (*gomcp.CallToolResult, any, error) {
		ts, err := svc.StopTimer(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(ts)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name: "update_timespan",
		Description: `Edit an existing timespan's start, end, tags and note by its id (from list_timespans).
		Tags and note replace the existing values; omit end to leave the timespan running.`,
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.UpdateTimespanInput) (*gomcp.CallToolResult, any, error) {
		ts, err := svc.UpdateTimespan(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(ts)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_timers",
		Description: "List the currently running timers (open timespans).",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.ListTimersInput) (*gomcp.CallToolResult, any, error) {
		timers, err := svc.ListTimers(ctx)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(timers)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name: "list_timespans",
		Description: `List recorded timespans, optionally bounded by an RFC3339 time range.
		Use this to find the id of a timespan to edit with update_timespan.`,
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.ListTimespansInput) (*gomcp.CallToolResult, any, error) {
		spans, err := svc.ListTimespans(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(spans)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_tags",
		Description: "List all tag definitions with their key, color and usage count.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.ListTagsInput) (*gomcp.CallToolResult, any, error) {
		tags, err := svc.ListTags(ctx)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(tags)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "create_tag",
		Description: "Create a new tag definition with a key and a hex color (e.g. #4caf50).",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.CreateTagInput) (*gomcp.CallToolResult, any, error) {
		tag, err := svc.CreateTag(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(tag)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "update_tag",
		Description: "Update a tag's color and optionally rename it via newKey.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.UpdateTagInput) (*gomcp.CallToolResult, any, error) {
		tag, err := svc.UpdateTag(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(tag)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "remove_tag",
		Description: "Remove a tag definition by its key.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.RemoveTagInput) (*gomcp.CallToolResult, any, error) {
		tag, err := svc.RemoveTag(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(tag)
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "remove_timespan",
		Description: "Remove a timespan by its id (from list_timespans).",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in traggo.RemoveTimespanInput) (*gomcp.CallToolResult, any, error) {
		ts, err := svc.RemoveTimespan(ctx, in)
		if err != nil {
			return nil, nil, err
		}
		return jsonResult(ts)
	})

	return &McpServer{
		Handler: gomcp.NewStreamableHTTPHandler(func(r *http.Request) *gomcp.Server {
			return server
		}, nil),
		srv: server,
	}

}
