package slack

import "time"

type EventType string

const (
	EventTypeMessage     EventType = "message"
	EventTypeSlashCommand EventType = "slash_command"
	EventTypeAppMention  EventType = "app_mention"
	EventTypeChannelJoin EventType = "channel_join"
	EventTypeReaction    EventType = "reaction"
)

type MessageEvent struct {
	EventType EventType `json:"event_type"`
	
	// Common fields
	TeamID    string `json:"team_id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Timestamp string `json:"timestamp"`
	
	// Message content
	Text      string `json:"text"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	
	// Command-specific fields (for slash commands)
	Command     string `json:"command,omitempty"`
	ResponseURL string `json:"response_url,omitempty"`
	TriggerID   string `json:"trigger_id,omitempty"`
	
	// Reaction-specific fields
	Reaction string `json:"reaction,omitempty"`
	
	// Raw event data for advanced processing
	RawEvent map[string]interface{} `json:"raw_event,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
}