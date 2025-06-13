package infragpt

import (
	"context"
)

type ConversationService interface {
	CompleteSlackIntegration(context.Context, CompleteSlackIntegrationCommand) error

	SendReply(context.Context, SendReplyCommand) error
}

type CompleteSlackIntegrationCommand struct {
	BusinessID string
	Code       string
}

type SendReplyCommand struct {
	ConversationID string
	Message        string
}
