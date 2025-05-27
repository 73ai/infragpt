package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// ActionPlan represents the plan to be executed based on user request
type ActionPlan struct {
	RequestID         string   `json:"request_id"`
	RequesterID       string   `json:"requester_id"`
	RequesterName     string   `json:"requester_name,omitempty"`
	RequesterUsername string   `json:"requester_username,omitempty"`
	ApproverID        string   `json:"approver_id"`
	ApproverName      string   `json:"approver_name,omitempty"`
	ApproverUsername  string   `json:"approver_username,omitempty"`
	Resource          string   `json:"resource"`
	Action            string   `json:"action"`
	Steps             []string `json:"steps"`
	Status            string   `json:"status"` // "pending", "approved", "rejected", "executed", "failed"
	ErrorMessage      string   `json:"error_message,omitempty"`
	Timestamp         string   `json:"timestamp"`
	ApprovalMsgTs     string   `json:"approval_msg_ts,omitempty"`
}

// InMemoryStorage for storing action plans (in a real app, use a database)
type InMemoryStorage struct {
	Plans map[string]*ActionPlan
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		Plans: make(map[string]*ActionPlan),
	}
}

func (s *InMemoryStorage) SavePlan(plan *ActionPlan) {
	s.Plans[plan.RequestID] = plan
}

func (s *InMemoryStorage) GetPlan(requestID string) (*ActionPlan, error) {
	plan, exists := s.Plans[requestID]
	if !exists {
		return nil, errors.New("plan not found")
	}
	return plan, nil
}

// AIService interface for analyzing user requests and generating action plans
type AIService interface {
	AnalyzeRequest(request string, requesterID string) (*ActionPlan, error)
}

// SimpleAIService is a mock implementation of AIService
type SimpleAIService struct{}

func NewSimpleAIService() *SimpleAIService {
	return &SimpleAIService{}
}

func (s *SimpleAIService) AnalyzeRequest(request string, requesterID string) (*ActionPlan, error) {
	// In a real app, this would call your InfraGPT AI service
	// For demo purposes, we'll just parse the request based on keywords

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	resource := "unknown"
	action := "unknown"

	// Simple parsing logic (replace with actual AI call in production)
	requestLower := strings.ToLower(request)

	if strings.Contains(requestLower, "s3") || strings.Contains(requestLower, "bucket") {
		resource = "s3-bucket"
	} else if strings.Contains(requestLower, "ec2") || strings.Contains(requestLower, "instance") {
		resource = "ec2-instance"
	} else if strings.Contains(requestLower, "lambda") || strings.Contains(requestLower, "function") {
		resource = "lambda-function"
	}

	if strings.Contains(requestLower, "access") || strings.Contains(requestLower, "permission") {
		action = "grant-access"
	} else if strings.Contains(requestLower, "create") {
		action = "create-resource"
	} else if strings.Contains(requestLower, "delete") {
		action = "delete-resource"
	}

	steps := []string{
		fmt.Sprintf("Identify the %s resource", resource),
		"Verify requester's current permissions",
		fmt.Sprintf("Perform %s operation", action),
		"Update permissions in IAM",
		"Verify changes were applied correctly",
	}

	return &ActionPlan{
		RequestID:   requestID,
		RequesterID: requesterID,
		Resource:    resource,
		Action:      action,
		Steps:       steps,
		Status:      "pending",
		Timestamp:   time.Now().Format(time.RFC3339),
	}, nil
}

// CloudService for executing action plans
type CloudService interface {
	ExecutePlan(plan *ActionPlan) error
}

// MockCloudService is a mock implementation of CloudService
type MockCloudService struct{}

func NewMockCloudService() *MockCloudService {
	return &MockCloudService{}
}

func (s *MockCloudService) ExecutePlan(plan *ActionPlan) error {
	// In a real app, this would actually perform the cloud operations
	time.Sleep(2 * time.Second) // Simulate work

	// For demo, randomly succeed or fail
	if time.Now().UnixNano()%10 == 0 {
		plan.Status = "failed"
		plan.ErrorMessage = "Simulated failure for demonstration purposes"
		return errors.New(plan.ErrorMessage)
	}

	plan.Status = "executed"
	return nil
}

// SlackBot handles Slack interactions
type SlackBot struct {
	client       *slack.Client
	socketClient *socketmode.Client
	aiService    AIService
	cloudService CloudService
	storage      *InMemoryStorage
	botUserID    string
}

func NewSlackBot(token string, appToken string, aiService AIService, cloudService CloudService, storage *InMemoryStorage) (*SlackBot, error) {
	client := slack.New(token, slack.OptionAppLevelToken(appToken))
	socketClient := socketmode.New(client)

	// Get bot user ID
	authTest, err := client.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("auth test failed: %w", err)
	}

	return &SlackBot{
		client:       client,
		socketClient: socketClient,
		aiService:    aiService,
		cloudService: cloudService,
		storage:      storage,
		botUserID:    authTest.UserID,
	}, nil
}

func (bot *SlackBot) Start(ctx context.Context) error {
	go bot.handleEvents(ctx)
	return bot.socketClient.Run()
}

func (bot *SlackBot) handleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-bot.socketClient.Events:
			log.Printf("Received event type: %s", event.Type)
			switch event.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack...")
			case socketmode.EventTypeConnectionError:
				log.Println("Connection error:", event.Data)
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack!")
			case socketmode.EventTypeEventsAPI:
				bot.socketClient.Ack(*event.Request)
				payload, _ := event.Data.(slackevents.EventsAPIEvent)
				log.Printf("Received Events API event: %s", payload.Type)
				bot.handleEventAPI(payload)
			case socketmode.EventTypeInteractive:
				bot.socketClient.Ack(*event.Request)
				payload, _ := event.Data.(slack.InteractionCallback)
				bot.handleInteraction(payload)
			default:
				log.Printf("Unhandled event type: %s with data: %v", event.Type, event.Data)
			}
		}
	}
}

func (bot *SlackBot) handleEventAPI(event slackevents.EventsAPIEvent) {
	log.Printf("Handling event API event of type: %s", event.Type)
	switch event.Type {
	case slackevents.CallbackEvent:
		log.Printf("Received callback event, inner event type: %s", event.InnerEvent.Type)
		switch ev := event.InnerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			log.Printf("Received app mention from user %s in channel %s with text: %s",
				ev.User, ev.Channel, ev.Text)
			bot.handleAppMention(ev)
		default:
			log.Printf("Unhandled inner event type: %T", event.InnerEvent.Data)
		}
	case slackevents.URLVerification:
		log.Printf("Received URL verification event")
	default:
		log.Printf("Unhandled Events API event type: %s", event.Type)
	}
}

func (bot *SlackBot) handleAppMention(event *slackevents.AppMentionEvent) {
	// Extract text without the bot mention
	text := strings.TrimSpace(strings.Replace(event.Text, fmt.Sprintf("<@%s>", bot.botUserID), "", -1))

	// Acknowledge the request in the same thread
	bot.client.PostMessage(
		event.Channel,
		slack.MsgOptionText(fmt.Sprintf("I'm analyzing your request: *%s*", text), false),
		slack.MsgOptionTS(event.TimeStamp), // This keeps replies in the thread
	)

	// Get requester info
	requesterInfo, err := bot.client.GetUserInfo(event.User)
	requesterName := ""
	requesterUsername := ""
	if err == nil && requesterInfo != nil {
		requesterName = requesterInfo.RealName
		requesterUsername = requesterInfo.Name // This is the @username
		log.Printf("Found requester: %s (@%s) for ID: %s", requesterName, requesterUsername, event.User)
	} else {
		log.Printf("Error getting requester info: %s", err.Error())
	}

	// Process the request with AI
	plan, err := bot.aiService.AnalyzeRequest(text, event.User)
	if err != nil {
		bot.client.PostMessage(event.Channel, slack.MsgOptionText(
			fmt.Sprintf("Error analyzing your request: %s", err.Error()),
			false,
		))
		return
	}

	// Store requester info
	plan.RequesterName = requesterName
	plan.RequesterUsername = requesterUsername

	// Get approver info directly from Slack API
	approverID := os.Getenv("DEFAULT_APPROVER_ID")
	if approverID == "" {
		log.Println("WARNING: DEFAULT_APPROVER_ID not set. Using a fallback ID.")
		approverID = "U0123456789" // Replace with a real admin user ID in production
	}

	// Get approver info using Slack API
	approverInfo, err := bot.client.GetUserInfo(approverID)
	if err != nil {
		log.Printf("Error getting approver info: %s. Will continue with ID only.", err.Error())
	} else if approverInfo != nil {
		// Store approver info from Slack
		plan.ApproverName = approverInfo.RealName
		plan.ApproverUsername = approverInfo.Name
		log.Printf("Found approver: %s (@%s) for ID: %s",
			plan.ApproverName, plan.ApproverUsername, approverID)
	}

	// Always set the approver ID
	plan.ApproverID = approverID

	// Save the plan
	bot.storage.SavePlan(plan)

	// Use Slack mentions format for users
	requesterDisplay := fmt.Sprintf("<@%s>", plan.RequesterID)
	approverDisplay := fmt.Sprintf("<@%s>", plan.ApproverID)

	// Send approval request to the approver with enhanced message indicating who can approve
	approvalText := fmt.Sprintf(
		"*New permission request from %s*\n\n"+
			"*Request:* %s\n"+
			"*Resource:* %s\n"+
			"*Action:* %s\n\n"+
			"*Proposed Action Plan:*\n",
		requesterDisplay,
		text,
		plan.Resource,
		plan.Action,
	)

	for i, step := range plan.Steps {
		approvalText += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	// Add clear indicator of who can approve
	approvalText += fmt.Sprintf("\n*⚠️ Only %s can approve or reject this request*", approverDisplay)

	// Create buttons for approval/rejection
	approveButton := slack.NewButtonBlockElement(
		fmt.Sprintf("approve_%s", plan.RequestID),
		plan.RequestID,
		slack.NewTextBlockObject("plain_text", "Approve", false, false),
	)
	approveButton.Style = slack.StylePrimary

	rejectButton := slack.NewButtonBlockElement(
		fmt.Sprintf("reject_%s", plan.RequestID),
		plan.RequestID,
		slack.NewTextBlockObject("plain_text", "Reject", false, false),
	)
	rejectButton.Style = slack.StyleDanger

	// Create action block with buttons
	actionBlock := slack.NewActionBlock(
		"approval_actions",
		approveButton,
		rejectButton,
	)

	// Send message to approver in the same thread as the original request
	_, timestamp, err := bot.client.PostMessage(
		event.Channel, // In a real app, use the approver's DM channel
		slack.MsgOptionText(approvalText, false),
		slack.MsgOptionTS(event.TimeStamp), // Keep in the same thread
		slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", approvalText, false, false),
				nil,
				nil,
			),
			actionBlock,
		),
	)

	if err != nil {
		log.Printf("Error sending approval request: %s", err.Error())
		bot.client.PostMessage(event.Channel, slack.MsgOptionText(
			"Error sending approval request. Please try again later.",
			false,
		))
		return
	}

	// Update plan with approval message timestamp
	plan.ApprovalMsgTs = timestamp
	bot.storage.SavePlan(plan)

	// Inform requester in the same thread - use proper Slack mention format
	bot.client.PostMessage(
		event.Channel,
		slack.MsgOptionText(
			fmt.Sprintf("I've sent your request for approval. You'll be notified once it's reviewed by %s.",
				approverDisplay),
			false,
		),
		slack.MsgOptionTS(event.TimeStamp), // Keep in the same thread
	)
}

func (bot *SlackBot) handleInteraction(interaction slack.InteractionCallback) {
	// Handle button clicks
	if interaction.Type == slack.InteractionTypeBlockActions {
		for _, action := range interaction.ActionCallback.BlockActions {
			if strings.HasPrefix(action.ActionID, "approve_") || strings.HasPrefix(action.ActionID, "reject_") {
				requestID := action.Value
				plan, err := bot.storage.GetPlan(requestID)
				if err != nil {
					log.Printf("Error retrieving plan: %s", err.Error())
					bot.respondToInteraction(interaction, "Error: Could not find the request details.")
					return
				}

				// Check if the user who clicked is the designated approver
				if interaction.User.ID != plan.ApproverID {
					log.Printf("Unauthorized approval attempt: User %s tried to act on request %s that should be approved by %s",
						interaction.User.ID, requestID, plan.ApproverID)

					// Use proper Slack mention format for approver ID
					approverDisplay := fmt.Sprintf("<@%s>", plan.ApproverID)

					bot.respondToInteraction(interaction,
						fmt.Sprintf("You are not authorized to approve or reject this request. Only %s can perform this action.",
							approverDisplay))
					return
				}

				isApproved := strings.HasPrefix(action.ActionID, "approve_")
				if isApproved {
					plan.Status = "approved"
					// Update the approval message with no buttons
					bot.updateApprovalMessage(interaction.Channel.ID, plan.ApprovalMsgTs, plan, "Approved! Executing plan...")

					// Execute the plan
					go bot.executeActionPlan(plan, interaction.Channel.ID)
				} else {
					plan.Status = "rejected"
					// Update the approval message with no buttons
					bot.updateApprovalMessage(interaction.Channel.ID, plan.ApprovalMsgTs, plan, "Rejected. No action taken.")

					// Use proper Slack mention format for requester ID
					requesterDisplay := fmt.Sprintf("<@%s>", plan.RequesterID)

					// Use proper Slack mention format for the user who clicked (approver)
					approverDisplay := fmt.Sprintf("<@%s>", interaction.User.ID)

					// Notify the requester in the same thread as approval message
					bot.client.PostMessage(
						interaction.Channel.ID,
						slack.MsgOptionText(
							fmt.Sprintf("%s: Your request has been rejected by %s.",
								requesterDisplay, approverDisplay),
							false,
						),
						slack.MsgOptionTS(plan.ApprovalMsgTs), // Keep in the same thread
					)
				}

				bot.storage.SavePlan(plan)
			}
		}
	}
}

func (bot *SlackBot) updateApprovalMessage(channelID string, timestamp string, plan *ActionPlan, statusText string) {
	// Use proper Slack mention format for requester ID
	requesterDisplay := fmt.Sprintf("<@%s>", plan.RequesterID)

	// Recreate the message text without buttons
	approvalText := fmt.Sprintf(
		"*Request from %s*\n\n"+
			"*Resource:* %s\n"+
			"*Action:* %s\n\n"+
			"*Action Plan:*\n",
		requesterDisplay,
		plan.Resource,
		plan.Action,
	)

	for i, step := range plan.Steps {
		approvalText += fmt.Sprintf("%d. %s\n", i+1, step)
	}

	// Use proper Slack mention format for approver ID
	approverDisplay := fmt.Sprintf("<@%s>", plan.ApproverID)

	// Add status text with approver information
	approvalText += fmt.Sprintf("\n*Status:* %s by %s", statusText, approverDisplay)

	// Add timestamp for audit purposes
	approvalText += fmt.Sprintf("\n*Time:* %s", time.Now().Format(time.RFC3339))

	// Create updated UI without action buttons
	_, _, _, _ = bot.client.UpdateMessage(
		channelID,
		timestamp,
		slack.MsgOptionText(approvalText, false),
		slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", approvalText, false, false),
				nil,
				nil,
			),
		),
	)
}

func (bot *SlackBot) executeActionPlan(plan *ActionPlan, channelID string) {
	// Execute the plan
	err := bot.cloudService.ExecutePlan(plan)
	statusText := ""

	if err != nil {
		statusText = fmt.Sprintf("*Execution failed:* %s", plan.ErrorMessage)
	} else {
		statusText = "*Execution successful!* All steps completed."
	}

	// Update the approval message
	bot.updateApprovalMessage(channelID, plan.ApprovalMsgTs, plan, statusText)

	// Use proper Slack mention format for requester and approver IDs
	requesterDisplay := fmt.Sprintf("<@%s>", plan.RequesterID)
	approverDisplay := fmt.Sprintf("<@%s>", plan.ApproverID)

	// Notify both requester and approver in the same thread
	statusMsg := "Completed successfully"
	if plan.Status != "executed" {
		statusMsg = fmt.Sprintf("Failed: %s", plan.ErrorMessage)
	}

	bot.client.PostMessage(
		channelID,
		slack.MsgOptionText(
			fmt.Sprintf("%s and %s: Action plan for request `%s` has been executed.\n*Status:* %s",
				requesterDisplay, approverDisplay, plan.RequestID, statusMsg),
			false,
		),
		slack.MsgOptionTS(plan.ApprovalMsgTs), // Keep in the same thread
	)
}

func (bot *SlackBot) respondToInteraction(interaction slack.InteractionCallback, message string) {
	// Create a more visually distinct ephemeral message
	ephemeralMsg := fmt.Sprintf("*Notice:* %s", message)

	bot.client.PostEphemeral(
		interaction.Channel.ID,
		interaction.User.ID,
		slack.MsgOptionText(ephemeralMsg, false),
	)
}

func main() {
	// Load environment variables
	godotenv.Load()

	// Get Slack API tokens from environment variables
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	slackAppToken := os.Getenv("SLACK_APP_TOKEN")

	if slackToken == "" || slackAppToken == "" {
		log.Fatal("SLACK_BOT_TOKEN and SLACK_APP_TOKEN must be set")
	}

	log.Printf("Using Bot Token starting with: %s...", slackToken[:10])
	log.Printf("Using App Token starting with: %s...", slackAppToken[:10])

	// Initialize services
	aiService := NewSimpleAIService()
	cloudService := NewMockCloudService()
	storage := NewInMemoryStorage()

	// Create and start the bot
	bot, err := NewSlackBot(slackToken, slackAppToken, aiService, cloudService, storage)
	if err != nil {
		log.Fatalf("Error creating Slack bot: %s", err.Error())
	}

	log.Println("Starting InfraGPT Slack bot...")
	ctx := context.Background()
	if err := bot.Start(ctx); err != nil {
		log.Fatalf("Error starting Slack bot: %s", err.Error())
	}
}
