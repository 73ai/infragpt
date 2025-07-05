package domain

import "context"

type AgentRequest struct {
	Conversation Conversation
	Message      Message
	PastMessages []Message
}

type AgentResponse struct {
	ResponseText string
	Success      bool
	ErrorMessage string
}

type AgentService interface {
	ProcessMessage(ctx context.Context, request AgentRequest) (AgentResponse, error)
}
