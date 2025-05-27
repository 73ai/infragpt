package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/infragptapi"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/postgresconfig"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/supporting/postgres"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/supporting/slack"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	config, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(fmt.Errorf("error reading config file: %w", err))
	}

	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal(config, &yamlMap); err != nil {
		log.Fatalf("Error unmarshalling YAML: %v", err)
	}

	type Config struct {
		Port     int                   `yaml:"port"`
		Slack    slack.Config          `mapstructure:"slack"`
		Database postgresconfig.Config `mapstructure:"database"`
	}

	var c Config
	if err := mapstructure.Decode(yamlMap, &c); err != nil {
		log.Fatalf("Error decoding config: %v", err)
	}

	slackConfig := c.Slack
	db, err := postgres.Config{Config: c.Database}.New()
	if err != nil {
		panic(fmt.Errorf("error connecting to database: %w", err))
	}
	slackConfig.WorkSpaceTokenRepository = db
	slackConfig.ChannelRepository = db

	sr, err := slackConfig.New(ctx)
	if err != nil {
		panic(fmt.Errorf("error connecting to slack: %w", err))
	}

	svcConfig := infragptsvc.Config{
		SlackGateway:             sr,
		IntegrationRepository:    db,
		ConversationRepository:   db,
		ChannelRepository:        db,
	}

	svc, err := svcConfig.New(ctx)
	if err != nil {
		panic(fmt.Errorf("error connecting to slack: %w", err))
	}

	g.Go(func() error {
		err = svc.SubscribeSlackNotifications(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			slog.Info("slack notification subscription stopped")
		}
		if err != nil {
			panic(fmt.Errorf("error subscribing to slack notifications: %w", err))
		}
		return nil
	})

	httpServer := &http.Server{
		Addr:        fmt.Sprintf(":%d", c.Port),
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     infragptapi.NewHandler(svc),
	}

	g.Go(func() error {
		slog.Info("infragpt: http server starting", "port", c.Port)
		err = httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("autopayd: http server stopped")
			return nil
		}
		slog.Error("autopayd: http server failed", "error", err)
		return fmt.Errorf("http server failed: %w", err)
	})

	if err := g.Wait(); err != nil {
		panic(fmt.Errorf("error waiting for server to finish: %w", err))
	}
}
