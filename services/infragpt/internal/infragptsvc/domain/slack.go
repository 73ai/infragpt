package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RequestApprovalCommand struct {
}

type Approval struct {
}

type Conversation struct {
	ID        uuid.UUID
	TeamID    string
	ChannelID string
	ThreadTS  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SlackMessageTS string
	Sender         SlackUser
	MessageText    string
	IsBotMessage   bool
	CreatedAt      time.Time
}

type Channel struct {
	ChannelID   string
	TeamID      string
	ChannelName string
	IsMonitored bool
	CreatedAt   time.Time
}

type SlackUser struct {
	ID       string
	Email    string
	Name     string
	Username string
}

type SlackThread struct {
	Message  string
	Sender   SlackUser
	Channel  string
	ThreadTS string
	TeamID   string
}

type MessageType string

const (
	MessageTypeAppMention MessageType = "app_mention"
	MessageTypeChannel    MessageType = "channel_message"
	MessageTypeThread     MessageType = "thread_message"
)

type UserCommand struct {
	Thread      SlackThread
	MessageTS   string
	InReply     bool
	MessageType MessageType
}

type SlackGateway interface {
	CompleteAuthentication(ctx context.Context, code string) (projectID string, err error)

	SubscribeAllMessages(context.Context, func(ctx context.Context, command UserCommand) error) error

	ReplyMessage(ctx context.Context, t SlackThread, message string) error
}

type WorkSpaceTokenRepository interface {
	SaveToken(ctx context.Context, teamID, token string) error
	GetToken(ctx context.Context, teamID string) (string, error)
}

type ConversationRepository interface {
	GetConversationByThread(ctx context.Context, teamID, channelID, threadTS string) (Conversation, error)
	CreateConversation(ctx context.Context, teamID, channelID, threadTS string) (Conversation, error)
	StoreMessage(ctx context.Context, conversationID uuid.UUID, message Message) (Message, error)
	MessageBySlackTS(ctx context.Context, conversationID uuid.UUID, senderID, slackMessageTS string) (Message, error)
	GetConversationHistory(ctx context.Context, conversationID uuid.UUID) ([]Message, error)
}

type ChannelRepository interface {
	AddChannel(ctx context.Context, teamID, channelID, channelName string) error
	SetChannelMonitoring(ctx context.Context, teamID, channelID string, isMonitored bool) error
	GetMonitoredChannels(ctx context.Context, teamID string) ([]Channel, error)
	IsChannelMonitored(ctx context.Context, teamID, channelID string) (bool, error)
}
