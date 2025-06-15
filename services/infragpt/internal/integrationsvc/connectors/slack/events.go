package slack

import "time"

type EventType string

const (
	EventTypeMessage      EventType = "message"
	EventTypeSlashCommand EventType = "slash_command"
	EventTypeAppMention   EventType = "app_mention"
	EventTypeChannelJoin  EventType = "channel_join"
	EventTypeReaction     EventType = "reaction"
)

type MessageEvent struct {
	EventType   EventType
	TeamID      string
	ChannelID   string
	UserID      string
	Timestamp   string
	Text        string
	ThreadTS    string
	Command     string
	ResponseURL string
	TriggerID   string
	Reaction    string
	RawEvent    map[string]any
	CreatedAt   time.Time
}
