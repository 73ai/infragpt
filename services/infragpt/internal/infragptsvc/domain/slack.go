package domain

import (
	"context"
)

type RequestApprovalCommand struct {
}

type Message struct {
}

type Approval struct {
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

type UserCommand struct {
	Thread  SlackThread
	InReply bool
}

type SlackGateway interface {
	CompleteAuthentication(ctx context.Context, code string) (projectID string, err error)

	SubscribeAppMentioned(context.Context, func(ctx context.Context, command UserCommand) error) error

	ReplyMessage(ctx context.Context, t SlackThread, message string) error
}

type WorkSpaceTokenRepository interface {
	SaveToken(ctx context.Context, teamID, token string) error
	GetToken(ctx context.Context, teamID string) (string, error)
}
