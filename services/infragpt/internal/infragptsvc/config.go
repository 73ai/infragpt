package infragptsvc

import (
	"context"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
)

type Config struct {
	SlackGateway          domain.SlackGateway
	IntegrationRepository domain.IntegrationRepository
}

func (c Config) New(ctx context.Context) (*Service, error) {
	if c.SlackGateway == nil {
		return nil, fmt.Errorf("slack gateway is required")
	}
	if c.IntegrationRepository == nil {
		return nil, fmt.Errorf("integration repository is required")
	}
	return &Service{
		slackGateway:          c.SlackGateway,
		integrationRepository: c.IntegrationRepository,
	}, nil
}
