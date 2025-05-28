package agent

import (
	"context"

	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
)

type DumbClient struct{}

func NewDumbClient() *DumbClient {
	return &DumbClient{}
}

func (c *DumbClient) ProcessMessage(ctx context.Context, request domain.AgentRequest) (domain.AgentResponse, error) {
	return domain.AgentResponse{
		ResponseText: "Hello, this is a dummy response from the agent.",
		Success:      true,
		ErrorMessage: "",
	}, nil
}
