package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/identityapi"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	agentclient "github.com/priyanshujain/infragpt/services/agent/src/client/go"
	"github.com/priyanshujain/infragpt/services/infragpt/infragptapi"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/httplog"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/generic/postgresconfig"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/domain"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/supporting/agent"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/supporting/postgres"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/infragptsvc/supporting/slack"
	"golang.org/x/sync/errgroup"

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
		LogLevel string                `mapstructure:"log_level"`
		Port     int                   `mapstructure:"port"`
		GrpcPort int                   `mapstructure:"grpc_port"`
		HttpLog  bool                  `mapstructure:"http_log"`
		Slack    slack.Config          `mapstructure:"slack"`
		Database postgresconfig.Config `mapstructure:"database"`
		Agent    agentclient.Config    `mapstructure:"agent"`
		Identity identitysvc.Config    `mapstructure:"identity"`
	}

	var c Config
	if err := mapstructure.Decode(yamlMap, &c); err != nil {
		log.Fatalf("Error decoding config: %v", err)
	}

	var level slog.Level
	// parse level from string
	if err := level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	
	slackConfig := c.Slack
	db, err := postgres.Config{Config: c.Database}.New()
	if err != nil {
		panic(fmt.Errorf("error connecting to database: %w", err))
	}
	slackConfig.WorkSpaceTokenRepository = db
	slackConfig.ChannelRepository = db

	// Create identity service with underlying database connection
	identityService := c.Identity.New(db.DB())

	authMiddleware := c.Identity.Clerk.NewAuthMiddleware()

	sr, err := slackConfig.New(ctx)
	if err != nil {
		panic(fmt.Errorf("error connecting to slack: %w", err))
	}

	// Create agent service with config from YAML
	var agentService domain.AgentService
	c.Agent.Timeout = 5 * 60 * time.Second
	c.Agent.ConnectTimeout = 10 * time.Second
	agentClient, err := agent.NewClient(&c.Agent)
	if err != nil {
		log.Printf("Failed to create agent client, falling back to DumbClient: %v", err)
	} else {
		agentService = agentClient
	}

	svcConfig := infragptsvc.Config{
		SlackGateway:           sr,
		IntegrationRepository:  db,
		ConversationRepository: db,
		ChannelRepository:      db,
		AgentService:           agentService,
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

	coreAPIHandler := infragptapi.NewHandler(svc)
	identityAPIHandler := identityapi.NewHandler(identityService, authMiddleware)

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				slog.Info("infragpt: http server panic", "recover", r)
			}
		}()
		if strings.HasPrefix(r.URL.Path, "/identity/") {
			identityAPIHandler.ServeHTTP(w, r)
			return
		}
		coreAPIHandler.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:        fmt.Sprintf(":%d", c.Port),
		BaseContext: func(net.Listener) context.Context { return ctx },
		Handler:     httplog.Middleware(c.HttpLog)(corsHandler(httpHandler)),
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

	grpcServer := infragptapi.NewGRPCServer(svc)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.GrpcPort))
	if err != nil {
		panic(fmt.Errorf("error creating grpc listener: %w", err))
	}

	g.Go(func() error {
		slog.Info("infragpt: grpc server starting", "port", c.GrpcPort)
		err = grpcServer.Serve(grpcListener)
		if err != nil {
			slog.Error("infragpt: grpc server failed", "error", err)
			return fmt.Errorf("grpc server failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		// run identity service webhook server
		slog.Info("infragpt: identity service webhook server starting", "port", c.Identity.Clerk.Port)
		err = identityService.Subscribe(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			slog.Info("infragpt: identity service webhook server stopped")
			return nil
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		panic(fmt.Errorf("error waiting for server to finish: %w", err))
	}
}

func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
