package traggo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	TraggoURL string
}

func NewService(traggoURL string) *Service {
	return &Service{TraggoURL: traggoURL}
}

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// execute runs a GraphQL query against Traggo using the per-request device
// token stored in the context by the MCP auth middleware. When out is non-nil
// the response "data" object is decoded into it.
func (s *Service) execute(ctx context.Context, query string, variables map[string]any, out any) error {
	token, _ := ctx.Value("traggo_token").(string)
	if token == "" {
		return fmt.Errorf("unauthorized: missing token in context")
	}

	payload, err := json.Marshal(graphqlRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.TraggoURL+"/graphql", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Traggo authenticates device tokens via the "traggo <token>" scheme.
	req.Header.Set("Authorization", "traggo "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	var env struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(env.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", env.Errors[0].Message)
	}

	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}

	return nil
}

const timeSpanFields = `id start end note tags { key value }`

// AddTimespan records a completed (or running, if End is empty) timespan.
func (s *Service) AddTimespan(ctx context.Context, in AddTimespanInput) (*TimeSpan, error) {
	query := `mutation($start: Time!, $end: Time, $tags: [InputTimeSpanTag!], $note: String!) {
		createTimeSpan(start: $start, end: $end, tags: $tags, note: $note) { ` + timeSpanFields + ` }
	}`
	vars := map[string]any{"start": in.Start, "note": in.Note, "tags": in.Tags}
	if in.End != "" {
		vars["end"] = in.End
	}

	var out struct {
		CreateTimeSpan TimeSpan `json:"createTimeSpan"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.CreateTimeSpan, nil
}

// StartTimer starts a running timespan (no end). Start defaults to now.
func (s *Service) StartTimer(ctx context.Context, in StartTimerInput) (*TimeSpan, error) {
	start := in.Start
	if start == "" {
		start = time.Now().Format(time.RFC3339)
	}

	query := `mutation($start: Time!, $tags: [InputTimeSpanTag!], $note: String!) {
		createTimeSpan(start: $start, tags: $tags, note: $note) { ` + timeSpanFields + ` }
	}`
	vars := map[string]any{"start": start, "note": in.Note, "tags": in.Tags}

	var out struct {
		CreateTimeSpan TimeSpan `json:"createTimeSpan"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.CreateTimeSpan, nil
}

// StopTimer closes a running timespan. End defaults to now.
func (s *Service) StopTimer(ctx context.Context, in StopTimerInput) (*TimeSpan, error) {
	end := in.End
	if end == "" {
		end = time.Now().Format(time.RFC3339)
	}

	query := `mutation($id: Int!, $end: Time!) {
		stopTimeSpan(id: $id, end: $end) { ` + timeSpanFields + ` }
	}`
	vars := map[string]any{"id": in.ID, "end": end}

	var out struct {
		StopTimeSpan TimeSpan `json:"stopTimeSpan"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.StopTimeSpan, nil
}

// UpdateTimespan edits an existing timespan's start, end, tags and note.
func (s *Service) UpdateTimespan(ctx context.Context, in UpdateTimespanInput) (*TimeSpan, error) {
	query := `mutation($id: Int!, $start: Time!, $end: Time, $tags: [InputTimeSpanTag!], $note: String!) {
		updateTimeSpan(id: $id, start: $start, end: $end, tags: $tags, note: $note) { ` + timeSpanFields + ` }
	}`
	vars := map[string]any{"id": in.ID, "start": in.Start, "note": in.Note, "tags": in.Tags}
	if in.End != "" {
		vars["end"] = in.End
	}

	var out struct {
		UpdateTimeSpan TimeSpan `json:"updateTimeSpan"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.UpdateTimeSpan, nil
}

// ListTimers returns the currently running timespans.
func (s *Service) ListTimers(ctx context.Context) ([]TimeSpan, error) {
	query := `query { timers { ` + timeSpanFields + ` } }`

	var out struct {
		Timers []TimeSpan `json:"timers"`
	}
	if err := s.execute(ctx, query, nil, &out); err != nil {
		return nil, err
	}
	return out.Timers, nil
}

// ListTimespans returns recorded timespans, optionally bounded by a time range.
func (s *Service) ListTimespans(ctx context.Context, in ListTimespansInput) ([]TimeSpan, error) {
	query := `query($from: Time, $to: Time) {
		timeSpans(fromInclusive: $from, toInclusive: $to) {
			timeSpans { ` + timeSpanFields + ` }
		}
	}`
	vars := map[string]any{}
	if in.FromInclusive != "" {
		vars["from"] = in.FromInclusive
	}
	if in.ToInclusive != "" {
		vars["to"] = in.ToInclusive
	}

	var out struct {
		TimeSpans struct {
			TimeSpans []TimeSpan `json:"timeSpans"`
		} `json:"timeSpans"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return out.TimeSpans.TimeSpans, nil
}

// RemoveTimespan deletes a timespan by its id.
func (s *Service) RemoveTimespan(ctx context.Context, in RemoveTimespanInput) (*TimeSpan, error) {
	query := `mutation($id: Int!) {
		removeTimeSpan(id: $id) { ` + timeSpanFields + ` }
	}`
	vars := map[string]any{"id": in.ID}

	var out struct {
		RemoveTimeSpan TimeSpan `json:"removeTimeSpan"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.RemoveTimeSpan, nil
}

// ListTags returns all tag definitions.
func (s *Service) ListTags(ctx context.Context) ([]TagDefinition, error) {
	query := `query { tags { key color usages } }`

	var out struct {
		Tags []TagDefinition `json:"tags"`
	}
	if err := s.execute(ctx, query, nil, &out); err != nil {
		return nil, err
	}
	return out.Tags, nil
}

// CreateTag creates a new tag definition.
func (s *Service) CreateTag(ctx context.Context, in CreateTagInput) (*TagDefinition, error) {
	query := `mutation($key: String!, $color: String!) {
		createTag(key: $key, color: $color) { key color usages }
	}`
	vars := map[string]any{"key": in.Key, "color": in.Color}

	var out struct {
		CreateTag TagDefinition `json:"createTag"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.CreateTag, nil
}

// UpdateTag changes a tag's color and optionally renames it (via newKey).
func (s *Service) UpdateTag(ctx context.Context, in UpdateTagInput) (*TagDefinition, error) {
	query := `mutation($key: String!, $newKey: String, $color: String!) {
		updateTag(key: $key, newKey: $newKey, color: $color) { key color usages }
	}`
	vars := map[string]any{"key": in.Key, "color": in.Color}
	if in.NewKey != "" {
		vars["newKey"] = in.NewKey
	}

	var out struct {
		UpdateTag TagDefinition `json:"updateTag"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.UpdateTag, nil
}

// RemoveTag deletes a tag definition by its key.
func (s *Service) RemoveTag(ctx context.Context, in RemoveTagInput) (*TagDefinition, error) {
	query := `mutation($key: String!) {
		removeTag(key: $key) { key color usages }
	}`
	vars := map[string]any{"key": in.Key}

	var out struct {
		RemoveTag TagDefinition `json:"removeTag"`
	}
	if err := s.execute(ctx, query, vars, &out); err != nil {
		return nil, err
	}
	return &out.RemoveTag, nil
}
