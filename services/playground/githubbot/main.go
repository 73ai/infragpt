package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v69/github"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Config holds application configuration
type Config struct {
	GithubAppID         int64
	GithubAppPrivateKey string // PEM encoded private key
	GithubAppName       string
	SlackBotToken       string
	SlackAppToken       string // App-level token for Socket Mode
	BaseURL             string // For callback URLs
	LogLevel            string // Added for configurable logging
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() Config {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Default log level
	}

	githubAppID := os.Getenv("GITHUB_APP_ID")
	appID := int64(0)
	if githubAppID != "" {
		fmt.Sscanf(githubAppID, "%d", &appID)
	}

	config := Config{
		GithubAppID:         appID,
		GithubAppPrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
		GithubAppName:       os.Getenv("GITHUB_APP_NAME"),
		SlackBotToken:       os.Getenv("SLACK_BOT_TOKEN"),
		SlackAppToken:       os.Getenv("SLACK_APP_TOKEN"),
		BaseURL:             os.Getenv("BASE_URL"),
		LogLevel:            logLevel,
	}

	// Validate required config
	if config.GithubAppID == 0 || config.GithubAppPrivateKey == "" || config.GithubAppName == "" {
		slog.Error("Missing GitHub App credentials in environment variables")
	}
	if config.SlackBotToken == "" || config.SlackAppToken == "" {
		slog.Error("Missing Slack credentials in environment variables")
	}
	if config.BaseURL == "" {
		slog.Error("Missing BASE_URL in environment variables")
	}

	slog.Info("Configuration loaded from environment variables",
		"github_app_id_available", config.GithubAppID != 0,
		"github_app_private_key_available", config.GithubAppPrivateKey != "",
		"github_app_name", config.GithubAppName,
		"slack_bot_token_available", config.SlackBotToken != "",
		"slack_app_token_available", config.SlackAppToken != "",
		"base_url", config.BaseURL,
		"log_level", config.LogLevel)

	return config
}

// SetupLogger configures the global logger
func SetupLogger(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logger initialized", "level", level)
}

// InstallationInfo stores GitHub App installation information
type InstallationInfo struct {
	InstallationID int64
	AccessToken    string
	TokenExpires   time.Time
}

// PendingAnalysis stores information about a GitHub link pending analysis
type PendingAnalysis struct {
	Link      GithubLinkInfo
	Channel   string
	User      string
	ThreadTS  string
	MessageTS string
}

// InstallationStore stores GitHub App installation information by repository
type InstallationStore struct {
	installations   sync.Map // "owner/repo" -> InstallationInfo
	pendingAnalyses sync.Map // messageTS -> []PendingAnalysis
}

// Store or update an installation for a repository
func (is *InstallationStore) StoreInstallation(ownerRepo string, info InstallationInfo) {
	is.installations.Store(ownerRepo, info)
	slog.Info("Stored installation for repository", "repo", ownerRepo, "installation_id", info.InstallationID)
}

// Get installation info for a repository
func (is *InstallationStore) GetInstallation(ownerRepo string) (InstallationInfo, bool) {
	value, exists := is.installations.Load(ownerRepo)
	if !exists {
		slog.Debug("Installation not found for repository", "repo", ownerRepo)
		return InstallationInfo{}, false
	}
	info, ok := value.(InstallationInfo)
	if !ok {
		slog.Error("Failed to cast installation info", "repo", ownerRepo)
		return InstallationInfo{}, false
	}
	return info, true
}

// Add a pending analysis for a message
func (is *InstallationStore) AddPendingAnalysis(messageTS string, analysis PendingAnalysis) {
	value, exists := is.pendingAnalyses.Load(messageTS)
	var analyses []PendingAnalysis
	if exists {
		analyses, _ = value.([]PendingAnalysis)
	}
	analyses = append(analyses, analysis)
	is.pendingAnalyses.Store(messageTS, analyses)
	slog.Info("Added pending analysis", "message_ts", messageTS, "repo", fmt.Sprintf("%s/%s", analysis.Link.Owner, analysis.Link.Repo))
}

// Get all pending analyses for a message
func (is *InstallationStore) GetPendingAnalyses(messageTS string) ([]PendingAnalysis, bool) {
	value, exists := is.pendingAnalyses.Load(messageTS)
	if !exists {
		return nil, false
	}
	analyses, ok := value.([]PendingAnalysis)
	if !ok {
		slog.Error("Failed to cast pending analyses", "message_ts", messageTS)
		return nil, false
	}
	return analyses, true
}

// Remove pending analyses for a message
func (is *InstallationStore) RemovePendingAnalyses(messageTS string) {
	is.pendingAnalyses.Delete(messageTS)
	slog.Info("Removed pending analyses", "message_ts", messageTS)
}

// GithubLinkInfo stores information extracted from a GitHub link
type GithubLinkInfo struct {
	Owner    string
	Repo     string
	IsFile   bool
	FilePath string
	Branch   string
	IsDir    bool
}

// Start the HTTP server and initialize handlers
func (app *App) Start() error {
	slog.Info("Starting application")

	// Set up HTTP routes for GitHub App callbacks
	http.HandleFunc("/github/webhook", app.handleGithubWebhook)
	http.HandleFunc("/github/callback", app.handleGithubCallback)
	slog.Info("HTTP routes configured")

	// Start the Socket Mode handler in a goroutine
	go app.startSocketModeHandler()

	serverAddr := ":8080"
	slog.Info("Starting HTTP server", "address", serverAddr)
	slog.Info("Starting Socket Mode client")

	return http.ListenAndServe(serverAddr, nil)
}

// Start the Socket Mode handler
func (app *App) startSocketModeHandler() {
	slog.Info("Initializing Socket Mode handler")

	// Configure the socket mode client
	go func() {
		for evt := range app.socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				slog.Info("Connecting to Slack with Socket Mode...")

			case socketmode.EventTypeConnectionError:
				slog.Error("Connection error with Slack", "error", evt.Data)

			case socketmode.EventTypeConnected:
				slog.Info("Connected to Slack with Socket Mode")

			case socketmode.EventTypeEventsAPI:
				// Get the event from the Envelope
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					slog.Error("Unexpected data type from Slack event", "type", fmt.Sprintf("%T", evt.Data))
					continue
				}

				// Log the event type
				slog.Debug("Received Slack event", "type", eventsAPIEvent.Type, "inner_event_type", eventsAPIEvent.InnerEvent.Type)

				// Acknowledge the event
				app.socketClient.Ack(*evt.Request)
				slog.Debug("Acknowledged Slack event")

				// Process the event
				app.processSlackEvent(eventsAPIEvent)

			case socketmode.EventTypeInteractive:
				// Handle interactive components if needed
				slog.Debug("Received interactive event", "data_type", fmt.Sprintf("%T", evt.Data))
				app.socketClient.Ack(*evt.Request)

			default:
				slog.Debug("Received unhandled event type", "type", evt.Type)
			}
		}
	}()

	// Start the client
	slog.Info("Starting Socket Mode client listener")
	err := app.socketClient.Run()
	if err != nil {
		slog.Error("Fatal error in Socket Mode client", "error", err)
		log.Fatal(err)
	}
}

// Process Slack events from Socket Mode
func (app *App) processSlackEvent(event slackevents.EventsAPIEvent) {
	// Only handle callback events
	if event.Type != slackevents.CallbackEvent {
		slog.Debug("Ignoring non-callback event", "event_type", event.Type)
		return
	}

	innerEvent := event.InnerEvent
	slog.Info("Processing Slack event", "inner_event_type", innerEvent.Type)

	switch innerEvent.Type {
	case string(slackevents.Message):
		// Process message events
		messageEvent, ok := innerEvent.Data.(*slackevents.MessageEvent)
		if !ok {
			slog.Error("Failed to cast to MessageEvent", "event_type", innerEvent.Type, "data_type", fmt.Sprintf("%T", innerEvent.Data))
			return
		}

		// Check if it's from a bot
		if messageEvent.BotID != "" {
			slog.Debug("Ignoring message from bot", "bot_id", messageEvent.BotID)
			return
		}

		// Process the message
		slog.Info("Processing user message", "channel", messageEvent.Channel, "user", messageEvent.User, "ts", messageEvent.TimeStamp)
		app.processMessage(messageEvent.Channel, messageEvent.User, messageEvent.Text, messageEvent.TimeStamp)

	case string(slackevents.AppMention):
		// Process app mention events
		appMentionEvent, ok := innerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			slog.Error("Failed to cast to AppMentionEvent", "event_type", innerEvent.Type, "data_type", fmt.Sprintf("%T", innerEvent.Data))
			return
		}

		// Process the mention - we'll use the same processMessage function
		slog.Info("Processing app mention", "channel", appMentionEvent.Channel, "user", appMentionEvent.User, "ts", appMentionEvent.TimeStamp)
		app.processMessage(appMentionEvent.Channel, appMentionEvent.User, appMentionEvent.Text, appMentionEvent.TimeStamp)

	default:
		slog.Debug("Unhandled inner event type", "type", innerEvent.Type)
	}
}

// Process a Slack message
func (app *App) processMessage(channel, user, text, threadTs string) {
	// Extract GitHub links
	slog.Debug("Analyzing message for GitHub links", "user", user, "channel", channel)
	links := extractGithubLinks(text)
	if len(links) == 0 {
		slog.Debug("No GitHub links found in message", "user", user, "channel", channel)
		return
	}
	slog.Info("GitHub links found in message", "count", len(links), "user", user, "channel", channel)

	// Group links by installation status
	var accessibleLinks, inaccessibleLinks []GithubLinkInfo

	for _, link := range links {
		repoFullName := fmt.Sprintf("%s/%s", link.Owner, link.Repo)

		// Check if we have an installation for this repository
		_, exists := app.installationStore.GetInstallation(repoFullName)
		if exists {
			accessibleLinks = append(accessibleLinks, link)
		} else {
			inaccessibleLinks = append(inaccessibleLinks, link)
		}
	}

	slog.Info("Links grouped by access",
		"accessible", len(accessibleLinks),
		"inaccessible", len(inaccessibleLinks),
		"user", user,
		"channel", channel)

	// Process accessible links immediately
	for _, link := range accessibleLinks {
		app.processGithubLink(channel, user, link, threadTs)
	}

	// If there are inaccessible links, provide installation instructions
	if len(inaccessibleLinks) > 0 {
		// Store the inaccessible links as pending analyses
		for _, link := range inaccessibleLinks {
			app.installationStore.AddPendingAnalysis(threadTs, PendingAnalysis{
				Link:      link,
				Channel:   channel,
				User:      user,
				ThreadTS:  threadTs,
				MessageTS: threadTs,
			})
		}

		// Create installation link message
		repoList := ""
		for i, link := range inaccessibleLinks {
			if i > 0 {
				repoList += ", "
			}
			repoList += fmt.Sprintf("%s/%s", link.Owner, link.Repo)
		}

		installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new", app.config.GithubAppName)

		slog.Info("Sending installation message to Slack", "url", installURL, "user", user, "channel", channel)

		message := fmt.Sprintf("I need access to analyze the following repositories: %s\n\nPlease install the GitHub App on these repositories: <%s|Install GitHub App>", repoList, installURL)

		_, _, err := app.slackClient.PostMessage(
			channel,
			slack.MsgOptionText(message, false),
			slack.MsgOptionTS(threadTs), // Reply in thread
		)
		if err != nil {
			slog.Error("Failed to send installation message to Slack", "error", err)
		}
	}
}

// Process a GitHub link
func (app *App) processGithubLink(channel, user string, link GithubLinkInfo, threadTs string) {
	repoFullName := fmt.Sprintf("%s/%s", link.Owner, link.Repo)
	slog.Info("Processing GitHub link",
		"user", user,
		"repo", repoFullName,
		"is_file", link.IsFile,
		"is_dir", link.IsDir)

	// Fetch content from GitHub
	content, err := app.fetchGithubContent(link)
	if err != nil {
		slog.Error("Error fetching GitHub content", "error", err, "user", user, "repo", repoFullName)

		_, _, slackErr := app.slackClient.PostMessage(
			channel,
			slack.MsgOptionText(fmt.Sprintf("Error fetching content: %v", err), false),
			slack.MsgOptionTS(threadTs), // Reply in thread
		)
		if slackErr != nil {
			slog.Error("Failed to send error message to Slack", "error", slackErr)
		}
		return
	}

	// Analyze content (mock implementation)
	slog.Info("Analyzing GitHub content", "user", user, "repo", repoFullName, "content_length", len(content))
	analysis := app.analyzeCode(content, link)

	// Send response in thread
	slog.Info("Sending analysis to Slack", "user", user, "channel", channel, "analysis_length", len(analysis))
	_, _, err = app.slackClient.PostMessage(
		channel,
		slack.MsgOptionText(analysis, false),
		slack.MsgOptionTS(threadTs), // Reply in thread
	)
	if err != nil {
		slog.Error("Failed to send analysis to Slack", "error", err)
	}
}

// Generate a JWT for GitHub App authentication
func (app *App) generateJWT() (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": app.config.GithubAppID,
	})

	tokenString, err := token.SignedString(app.privateKey)
	if err != nil {
		slog.Error("Failed to sign JWT", "error", err)
		return "", err
	}

	return tokenString, nil
}

// Get an installation token for a repository
func (app *App) getInstallationToken(ownerRepo string) (string, error) {
	// Check if we have a valid token already
	info, exists := app.installationStore.GetInstallation(ownerRepo)
	if exists && info.AccessToken != "" && time.Now().Before(info.TokenExpires) {
		slog.Debug("Using cached installation token", "repo", ownerRepo, "expires", info.TokenExpires)
		return info.AccessToken, nil
	}

	// No valid token, need to generate a new one
	slog.Info("Generating new installation token", "repo", ownerRepo, "installation_id", info.InstallationID)

	// Generate JWT for GitHub App authentication
	jwt, err := app.generateJWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %v", err)
	}

	// Request a new installation token
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", info.InstallationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		slog.Error("Failed to create token request", "error", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to request installation token", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("GitHub API error when requesting token",
			"status", resp.StatusCode,
			"response", string(body))
		return "", fmt.Errorf("failed to get installation token, status: %d", resp.StatusCode)
	}

	var tokenResponse struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		slog.Error("Failed to decode token response", "error", err)
		return "", err
	}

	// Update the stored installation info with the new token
	info.AccessToken = tokenResponse.Token
	info.TokenExpires = tokenResponse.ExpiresAt
	app.installationStore.StoreInstallation(ownerRepo, info)

	slog.Info("Generated new installation token", "repo", ownerRepo, "expires", tokenResponse.ExpiresAt)
	return tokenResponse.Token, nil
}

// Fix for the URL encoding issue in fetchGithubContent method
func (app *App) fetchGithubContent(info GithubLinkInfo) (string, error) {
	startTime := time.Now()
	repoFullName := fmt.Sprintf("%s/%s", info.Owner, info.Repo)
	slog.Info("Fetching GitHub content",
		"owner", info.Owner,
		"repo", info.Repo,
		"is_file", info.IsFile,
		"path", info.FilePath,
		"branch", info.Branch)

	// Get installation token
	token, err := app.getInstallationToken(repoFullName)
	if err != nil {
		slog.Error("Failed to get installation token", "error", err)
		return "", fmt.Errorf("failed to get access: %v", err)
	}

	// Create GitHub client with the installation token
	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()

	// If it's a file, fetch the file content
	if info.IsFile {
		slog.Debug("Fetching file content",
			"owner", info.Owner,
			"repo", info.Repo,
			"path", info.FilePath,
			"branch", info.Branch)

		// Clean the path - sometimes paths can have URL encoding or special characters
		cleanPath := strings.TrimSuffix(info.FilePath, ">") // Remove trailing '>' character if present
		cleanPath = strings.TrimSpace(cleanPath)            // Remove any whitespace

		slog.Debug("Using cleaned file path", "original", info.FilePath, "cleaned", cleanPath)

		fileContent, _, _, err := client.Repositories.GetContents(
			ctx,
			info.Owner,
			info.Repo,
			cleanPath,
			&github.RepositoryContentGetOptions{Ref: info.Branch},
		)
		if err != nil {
			slog.Error("Failed to get file from GitHub",
				"error", err,
				"path", cleanPath,
				"branch", info.Branch,
				"url", fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", info.Owner, info.Repo, info.Branch, cleanPath))

			// Attempt to get without branch specification if that fails
			if info.Branch != "main" && info.Branch != "master" {
				slog.Info("Trying to fetch with default branch instead", "current_branch", info.Branch)
				fileContent, _, _, err = client.Repositories.GetContents(
					ctx,
					info.Owner,
					info.Repo,
					cleanPath,
					nil, // No branch specified, use default
				)
				if err != nil {
					return "", fmt.Errorf("failed to get file: %v", err)
				}
			} else {
				return "", fmt.Errorf("failed to get file: %v", err)
			}
		}

		// Check file size (skip large files)
		if fileContent.Size != nil && *fileContent.Size > 1024*1024 {
			slog.Warn("File too large to analyze", "size", *fileContent.Size, "path", cleanPath)
			return "", fmt.Errorf("file too large to analyze (> 1MB)")
		}

		// Check content type (skip binary files)
		if fileContent.Type != nil && *fileContent.Type != "file" {
			slog.Warn("Not a text file", "type", *fileContent.Type, "path", cleanPath)
			return "", fmt.Errorf("not a text file")
		}

		// Decode content
		content, err := fileContent.GetContent()
		if err != nil {
			slog.Error("Failed to decode content from GitHub", "error", err)
			return "", fmt.Errorf("failed to decode content: %v", err)
		}

		elapsed := time.Since(startTime)
		slog.Info("Successfully fetched file content",
			"path", cleanPath,
			"content_length", len(content),
			"duration_ms", elapsed.Milliseconds())
		return content, nil
	}

	// For directories or repositories, fetch structure
	if info.IsDir {
		slog.Debug("Fetching directory structure", "owner", info.Owner, "repo", info.Repo, "path", info.FilePath)

		ctx := context.Background()
		_, contents, _, err := client.Repositories.GetContents(
			ctx,
			info.Owner,
			info.Repo,
			info.FilePath,
			&github.RepositoryContentGetOptions{Ref: info.Branch},
		)
		if err != nil {
			slog.Error("Failed to get directory contents from GitHub", "error", err)
			return "", fmt.Errorf("failed to get directory contents: %v", err)
		}

		// Build a string representing the directory structure
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Directory: %s/%s/%s\n\n", info.Owner, info.Repo, info.FilePath))
		for _, item := range contents {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", *item.Name, *item.Type))
		}

		content := sb.String()
		elapsed := time.Since(startTime)
		slog.Info("Successfully fetched directory structure", "path", info.FilePath, "item_count", len(contents), "duration_ms", elapsed.Milliseconds())
		return content, nil
	}

	// For repository root, fetch README if available
	slog.Debug("Fetching repository information", "owner", info.Owner, "repo", info.Repo)

	readme, _, err := client.Repositories.GetReadme(ctx, info.Owner, info.Repo, &github.RepositoryContentGetOptions{})
	if err != nil {
		// If README not found, return repo info
		slog.Info("README not found, fetching repository metadata", "error", err)

		repo, _, err := client.Repositories.Get(ctx, info.Owner, info.Repo)
		if err != nil {
			slog.Error("Failed to get repository info from GitHub", "error", err)
			return "", fmt.Errorf("failed to get repository info: %v", err)
		}

		repoInfo := fmt.Sprintf("Repository: %s\nDescription: %s\nStars: %d\nLanguage: %s",
			*repo.FullName,
			safeGetString(repo.Description),
			safeGetInt(repo.StargazersCount),
			safeGetString(repo.Language))

		elapsed := time.Since(startTime)
		slog.Info("Successfully fetched repository metadata", "repo", *repo.FullName, "duration_ms", elapsed.Milliseconds())
		return repoInfo, nil
	}

	// Decode README content
	content, err := readme.GetContent()
	if err != nil {
		slog.Error("Failed to decode README content from GitHub", "error", err)
		return "", fmt.Errorf("failed to decode README: %v", err)
	}

	elapsed := time.Since(startTime)
	slog.Info("Successfully fetched README", "repo", fmt.Sprintf("%s/%s", info.Owner, info.Repo), "content_length", len(content), "duration_ms", elapsed.Milliseconds())
	return fmt.Sprintf("README for %s/%s:\n\n%s", info.Owner, info.Repo, content), nil
}

// Handle GitHub webhook events
func (app *App) handleGithubWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify the webhook payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read webhook payload", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the webhook event
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		slog.Error("Failed to parse webhook", "error", err)
		http.Error(w, "Failed to parse webhook", http.StatusBadRequest)
		return
	}

	// Handle different webhook events
	switch e := event.(type) {
	case *github.InstallationEvent:
		if e.GetAction() == "created" || e.GetAction() == "added" {
			app.handleInstallationChange(e)
		}
	case *github.InstallationRepositoriesEvent:
		if e.GetAction() == "added" {
			app.handleInstallationRepositoriesChange(e)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Process any pending analyses after a new installation
func (app *App) processPendingAnalyses() {
	slog.Info("Processing pending analyses after installation")

	// Create a temporary map to track which message timestamps we've processed
	processedMessages := make(map[string]bool)

	// Iterate through all pending analyses
	app.installationStore.pendingAnalyses.Range(func(key, value interface{}) bool {
		messageTS, ok := key.(string)
		if !ok {
			slog.Error("Invalid message timestamp type", "type", fmt.Sprintf("%T", key))
			return true // continue iteration
		}

		analyses, ok := value.([]PendingAnalysis)
		if !ok {
			slog.Error("Invalid pending analyses type", "type", fmt.Sprintf("%T", value))
			return true // continue iteration
		}

		// Check each pending analysis
		var remainingAnalyses []PendingAnalysis
		var processedCount int

		for _, analysis := range analyses {
			repoFullName := fmt.Sprintf("%s/%s", analysis.Link.Owner, analysis.Link.Repo)

			// Check if we now have an installation for this repository
			_, exists := app.installationStore.GetInstallation(repoFullName)
			if exists {
				// Process this link now that we have access
				slog.Info("Processing previously pending link",
					"repo", repoFullName,
					"channel", analysis.Channel,
					"message_ts", messageTS)

				// Process the link
				go app.processGithubLink(analysis.Channel, analysis.User, analysis.Link, analysis.ThreadTS)
				processedCount++
			} else {
				// Still don't have access, keep it in the pending list
				remainingAnalyses = append(remainingAnalyses, analysis)
			}
		}

		// Update the pending analyses list
		if len(remainingAnalyses) == 0 {
			// All analyses for this message were processed
			processedMessages[messageTS] = true
			slog.Info("All pending analyses processed for message", "message_ts", messageTS)

			// Notify the user that we've processed all links
			if processedCount > 0 {
				go app.slackClient.PostMessage(
					analyses[0].Channel, // Use the channel from any analysis
					slack.MsgOptionText("✅ All GitHub links have been analyzed!", false),
					slack.MsgOptionTS(analyses[0].ThreadTS), // Reply in thread
				)
			}
		} else {
			// Some analyses still pending
			app.installationStore.pendingAnalyses.Store(messageTS, remainingAnalyses)
			slog.Info("Some analyses still pending for message",
				"message_ts", messageTS,
				"processed", processedCount,
				"remaining", len(remainingAnalyses))
		}

		return true // continue iteration
	})

	// Clean up fully processed messages
	for messageTS := range processedMessages {
		app.installationStore.pendingAnalyses.Delete(messageTS)
	}
}

// Handle GitHub app callback (this is for the OAuth flow completion)
func (app *App) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	// For simplicity, we're not implementing the full callback handler
	// This would typically handle the redirect after installation
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "GitHub App installation complete. You can close this window and return to Slack.")
}

// Mock implementation of code analysis
func (app *App) analyzeCode(content string, info GithubLinkInfo) string {
	startTime := time.Now()
	slog.Info("Analyzing code",
		"owner", info.Owner,
		"repo", info.Repo,
		"is_file", info.IsFile,
		"is_dir", info.IsDir,
		"path", info.FilePath,
		"content_length", len(content),
	)

	// In a real implementation, this would call your AI service
	// For now, we'll return a simple mock response
	var result string

	if info.IsFile {
		lineCount := len(strings.Split(content, "\n"))
		fileType := getFileType(info.FilePath)
		slog.Info("File analysis", "path", info.FilePath, "line_count", lineCount, "file_type", fileType)

		result = fmt.Sprintf("Analysis of %s/%s/%s:\n\n"+
			"This file contains %d lines of code.\n"+
			"It appears to be a %s file based on the extension.\n\n"+
			"Summary: This is a mock analysis of the code. In a real implementation, "+
			"this would be replaced with actual AI-generated analysis.",
			info.Owner, info.Repo, info.FilePath,
			lineCount,
			fileType)

		// add first 5 lines of code to the result
		lines := strings.Split(content, "\n")
		for i := 0; i < 5 && i < len(lines); i++ {
			result += fmt.Sprintf("\n\n%s", lines[i])
		}
	} else if info.IsDir {
		slog.Info("Directory analysis", "path", info.FilePath)

		result = fmt.Sprintf("Analysis of directory %s/%s/%s:\n\n"+
			"This directory contains multiple files as shown in the structure.\n\n"+
			"Summary: This is a mock analysis of the directory. In a real implementation, "+
			"this would be replaced with actual AI-generated analysis of the directory structure.",
			info.Owner, info.Repo, info.FilePath)
	} else {
		slog.Info("Repository analysis", "repo", fmt.Sprintf("%s/%s", info.Owner, info.Repo))

		result = fmt.Sprintf("Analysis of repository %s/%s:\n\n"+
			"The repository appears to contain project documentation and code.\n\n"+
			"Summary: This is a mock analysis of the repository. In a real implementation, "+
			"this would be replaced with actual AI-generated repository analysis.",
			info.Owner, info.Repo)
	}

	elapsed := time.Since(startTime)
	slog.Info("Analysis completed", "duration_ms", elapsed.Milliseconds())
	return result
}

// Extract GitHub links from a message
func extractGithubLinks(text string) []GithubLinkInfo {
	// Regular expression to match GitHub links
	re := regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)(?:/blob/([^/]+))?/?(.*)`)

	// Find all matches in the text
	matches := re.FindAllStringSubmatch(text, -1)

	// Extract link information
	var links []GithubLinkInfo
	for _, match := range matches {
		link := GithubLinkInfo{
			Owner:    match[1],
			Repo:     match[2],
			IsFile:   match[3] != "",
			FilePath: match[4],
			Branch:   "main", // Default branch
		}

		if link.IsFile {
			link.Branch = match[3]
		}

		// Check if it's a directory
		if strings.HasSuffix(link.FilePath, "/") {
			link.IsDir = true
			link.FilePath = strings.TrimSuffix(link.FilePath, "/")
		}

		links = append(links, link)
	}

	return links
}

// Get the file type based on the file extension
func getFileType(filePath string) string {
	// Mock implementation - just check the file extension
	if strings.HasSuffix(filePath, ".go") {
		return "Go"
	} else if strings.HasSuffix(filePath, ".py") {
		return "Python"
	} else if strings.HasSuffix(filePath, ".js") {
		return "JavaScript"
	} else if strings.HasSuffix(filePath, ".md") {
		return "Markdown"
	}

	return "text"
}

// Safely get a string pointer value
func safeGetString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Safely get an int pointer value
func safeGetInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// Persistent storage implementation using JSON files
type Storage interface {
	SaveInstallations() error
	LoadInstallations() error
}

// Implementation of persistent storage using JSON files
type FileStorage struct {
	installationStore *InstallationStore
	filePath          string
}

// Create a new file storage instance
func NewFileStorage(installationStore *InstallationStore, filePath string) *FileStorage {
	return &FileStorage{
		installationStore: installationStore,
		filePath:          filePath,
	}
}

// StoredInstallation is a serializable version of InstallationInfo
type StoredInstallation struct {
	Repository     string    `json:"repository"`
	InstallationID int64     `json:"installation_id"`
	AccessToken    string    `json:"access_token,omitempty"`
	TokenExpires   time.Time `json:"token_expires,omitempty"`
}

// Save installations to a file
func (fs *FileStorage) SaveInstallations() error {
	var installations []StoredInstallation

	// Convert sync.Map to slice for JSON serialization
	fs.installationStore.installations.Range(func(key, value interface{}) bool {
		repo, ok := key.(string)
		if !ok {
			slog.Error("Invalid repository key type", "type", fmt.Sprintf("%T", key))
			return true // continue iteration
		}

		info, ok := value.(InstallationInfo)
		if !ok {
			slog.Error("Invalid installation info type", "type", fmt.Sprintf("%T", value))
			return true // continue iteration
		}

		installations = append(installations, StoredInstallation{
			Repository:     repo,
			InstallationID: info.InstallationID,
			AccessToken:    info.AccessToken,
			TokenExpires:   info.TokenExpires,
		})

		return true // continue iteration
	})

	// Create JSON data
	data, err := json.MarshalIndent(installations, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal installations", "error", err)
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Failed to create directory", "dir", dir, "error", err)
		return err
	}

	// Write to file
	if err := os.WriteFile(fs.filePath, data, 0644); err != nil {
		slog.Error("Failed to write installations to file", "file", fs.filePath, "error", err)
		return err
	}

	slog.Info("Saved installations to file", "file", fs.filePath, "count", len(installations))
	return nil
}

// Load installations from a file
func (fs *FileStorage) LoadInstallations() error {
	// Check if file exists
	_, err := os.Stat(fs.filePath)
	if os.IsNotExist(err) {
		slog.Info("Installations file does not exist, starting with empty state", "file", fs.filePath)
		return nil // Not an error, just an empty state
	}

	// Read file
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		slog.Error("Failed to read installations file", "file", fs.filePath, "error", err)
		return err
	}

	// Parse JSON
	var installations []StoredInstallation
	if err := json.Unmarshal(data, &installations); err != nil {
		slog.Error("Failed to unmarshal installations", "error", err)
		return err
	}

	// Load into sync.Map
	for _, inst := range installations {
		info := InstallationInfo{
			InstallationID: inst.InstallationID,
			AccessToken:    inst.AccessToken,
			TokenExpires:   inst.TokenExpires,
		}
		fs.installationStore.StoreInstallation(inst.Repository, info)
	}

	slog.Info("Loaded installations from file", "file", fs.filePath, "count", len(installations))
	return nil
}

// Modifications to the App struct to include storage
type App struct {
	config            Config
	installationStore *InstallationStore
	slackClient       *slack.Client
	socketClient      *socketmode.Client
	privateKey        *rsa.PrivateKey
	storage           Storage
}

// Save installations before shutdown
func (app *App) SaveInstallations() {
	if app.storage != nil {
		if err := app.storage.SaveInstallations(); err != nil {
			slog.Error("Failed to save installations", "error", err)
		}
	}
}

// Initialize storage in the NewApp function
func NewApp(config Config) (*App, error) {
	slog.Info("Creating new application instance")

	// Parse private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(config.GithubAppPrivateKey))
	if err != nil {
		slog.Error("Failed to parse GitHub App private key", "error", err)
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	slog.Info("GitHub App private key parsed successfully")

	// Set up Slack client
	slackClient := slack.New(
		config.SlackBotToken,
		slack.OptionAppLevelToken(config.SlackAppToken),
		slack.OptionLog(log.New(os.Stdout, "slack-api: ", log.Lshortfile|log.LstdFlags)),
	)
	slog.Info("Slack client created")

	// Create Socket Mode client
	socketClient := socketmode.New(
		slackClient,
		socketmode.OptionDebug(config.LogLevel == "debug"),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	slog.Info("Socket Mode client created", "debug_mode", config.LogLevel == "debug")

	// Create installation store
	installationStore := &InstallationStore{}

	// Create storage
	storageDir := "./data"
	storage := NewFileStorage(installationStore, filepath.Join(storageDir, "installations.json"))

	// Load installations from file
	if err := storage.LoadInstallations(); err != nil {
		slog.Warn("Failed to load installations, starting with empty state", "error", err)
	}

	return &App{
		config:            config,
		installationStore: installationStore,
		slackClient:       slackClient,
		socketClient:      socketClient,
		privateKey:        privateKey,
		storage:           storage,
	}, nil
}

// Add a shutdown handler to save installations
func setupShutdownHandler(app *App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		slog.Info("Shutting down gracefully...")
		app.SaveInstallations()
		slog.Info("Shutdown complete")
		os.Exit(0)
	}()
}

// Add this to main function
func main() {
	// Setup logger first to ensure all operations are logged
	SetupLogger(os.Getenv("LOG_LEVEL"))
	slog.Info("Starting GitHub Integration Chatbot")

	// Load configuration from environment variables
	config := LoadConfigFromEnv()

	// Validate essential configuration
	if config.SlackBotToken == "" || config.SlackAppToken == "" ||
		config.GithubAppID == 0 || config.GithubAppPrivateKey == "" ||
		config.GithubAppName == "" || config.BaseURL == "" {
		slog.Error("Missing required configuration. Please set the following environment variables: SLACK_BOT_TOKEN, SLACK_APP_TOKEN, GITHUB_APP_ID, GITHUB_APP_PRIVATE_KEY, GITHUB_APP_NAME, BASE_URL")
		os.Exit(1)
	}

	slog.Info("Configuration loaded successfully")

	// Create application instance
	app, err := NewApp(config)
	if err != nil {
		slog.Error("Failed to create application", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown handler to save installations
	setupShutdownHandler(app)

	slog.Info("Application instance created, starting server")

	// Start the application
	if err := app.Start(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

// Add save on installation changes
func (app *App) handleInstallationChange(event *github.InstallationEvent) {
	slog.Info("Processing installation event",
		"action", event.GetAction(),
		"installation_id", event.GetInstallation().GetID())

	// Store installation information for each repository
	for _, repo := range event.Repositories {
		repoName := *repo.FullName
		info := InstallationInfo{
			InstallationID: event.GetInstallation().GetID(),
			// The token will be generated when needed
		}
		app.installationStore.StoreInstallation(repoName, info)
		slog.Info("Added installation for repository", "repo", repoName, "installation_id", info.InstallationID)
	}

	// Save installations to persistent storage
	if app.storage != nil {
		if err := app.storage.SaveInstallations(); err != nil {
			slog.Error("Failed to save installations after change", "error", err)
		}
	}

	// Process any pending analyses for these repositories
	app.processPendingAnalyses()
}

// Add save on installation repositories changes
func (app *App) handleInstallationRepositoriesChange(event *github.InstallationRepositoriesEvent) {
	slog.Info("Processing installation repositories event",
		"action", event.GetAction(),
		"installation_id", event.GetInstallation().GetID())

	// Store installation information for each repository
	for _, repo := range event.RepositoriesAdded {
		repoName := *repo.FullName
		info := InstallationInfo{
			InstallationID: event.GetInstallation().GetID(),
			// The token will be generated when needed
		}
		app.installationStore.StoreInstallation(repoName, info)
		slog.Info("Added installation for repository", "repo", repoName, "installation_id", info.InstallationID)
	}

	// Save installations to persistent storage
	if app.storage != nil {
		if err := app.storage.SaveInstallations(); err != nil {
			slog.Error("Failed to save installations after change", "error", err)
		}
	}

	// Process any pending analyses for these repositories
	app.processPendingAnalyses()
}
