package traggo

type Tag struct {
	Key   string `json:"key" jsonschema:"the tag key"`
	Value string `json:"value" jsonschema:"the tag value"`
}

type AddTimespanInput struct {
	Start string `json:"start" jsonschema:"RFC3339 start time, e.g. 2026-06-28T09:00:00Z"`
	End   string `json:"end,omitempty" jsonschema:"RFC3339 end time; omit for a running timespan"`
	Note  string `json:"note,omitempty" jsonschema:"free-text note"`
	Tags  []Tag  `json:"tags,omitempty" jsonschema:"list of key/value tags"`
}

type StartTimerInput struct {
	Start string `json:"start,omitempty" jsonschema:"RFC3339 start time; omit to start the timer now"`
	Note  string `json:"note,omitempty" jsonschema:"free-text note"`
	Tags  []Tag  `json:"tags,omitempty" jsonschema:"list of key/value tags to attach to the running timer"`
}

type StopTimerInput struct {
	ID  int    `json:"id" jsonschema:"id of the running timespan to stop (from list_timers)"`
	End string `json:"end,omitempty" jsonschema:"RFC3339 end time; omit to stop the timer now"`
}

type UpdateTimespanInput struct {
	ID    int    `json:"id" jsonschema:"id of the timespan to edit (from list_timespans)"`
	Start string `json:"start" jsonschema:"RFC3339 start time"`
	End   string `json:"end,omitempty" jsonschema:"RFC3339 end time; omit to leave the timespan running"`
	Note  string `json:"note,omitempty" jsonschema:"free-text note; replaces the existing note"`
	Tags  []Tag  `json:"tags,omitempty" jsonschema:"list of key/value tags; replaces the existing tags"`
}

type CreateTagInput struct {
	Key   string `json:"key" jsonschema:"the tag key, e.g. project or client"`
	Color string `json:"color" jsonschema:"hex color for the tag, e.g. #4caf50"`
}

type UpdateTagInput struct {
	Key    string `json:"key" jsonschema:"the existing tag key to update"`
	NewKey string `json:"newKey,omitempty" jsonschema:"new key to rename the tag to; omit to keep the current key"`
	Color  string `json:"color" jsonschema:"hex color for the tag, e.g. #4caf50"`
}

type RemoveTagInput struct {
	Key string `json:"key" jsonschema:"the tag key to remove"`
}

type RemoveTimespanInput struct {
	ID int `json:"id" jsonschema:"id of the timespan to remove (from list_timespans)"`
}

type ListTagsInput struct{}

type ListTimersInput struct{}

type ListTimespansInput struct {
	FromInclusive string `json:"fromInclusive,omitempty" jsonschema:"RFC3339 lower bound (inclusive) to filter timespans"`
	ToInclusive   string `json:"toInclusive,omitempty" jsonschema:"RFC3339 upper bound (inclusive) to filter timespans"`
}

// TimeSpan mirrors the Traggo TimeSpan type. End is empty for a running timer.
type TimeSpan struct {
	ID    int    `json:"id"`
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
	Note  string `json:"note"`
	Tags  []Tag  `json:"tags,omitempty"`
}

// TagDefinition mirrors the Traggo TagDefinition type.
type TagDefinition struct {
	Key    string `json:"key"`
	Color  string `json:"color"`
	Usages int    `json:"usages"`
}
