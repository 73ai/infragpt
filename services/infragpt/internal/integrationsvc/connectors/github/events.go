package github

import "time"

type EventType string

const (
	EventTypePush         EventType = "push"
	EventTypePullRequest  EventType = "pull_request"
	EventTypeInstallation EventType = "installation"
	EventTypeIssues       EventType = "issues"
	EventTypeRelease      EventType = "release"
	EventTypeWorkflowRun  EventType = "workflow_run"
)

type WebhookEvent struct {
	EventType EventType `json:"event_type"`
	
	// Common fields
	InstallationID int64  `json:"installation_id"`
	RepositoryID   int64  `json:"repository_id,omitempty"`
	RepositoryName string `json:"repository_name,omitempty"`
	SenderID       int64  `json:"sender_id"`
	SenderLogin    string `json:"sender_login"`
	
	// Action for events that have it (pull_request, issues, etc.)
	Action string `json:"action,omitempty"`
	
	// Common payload fields
	Ref      string `json:"ref,omitempty"`        // For push events
	Branch   string `json:"branch,omitempty"`     // Extracted from ref
	CommitSHA string `json:"commit_sha,omitempty"` // For push/PR events
	
	// Pull Request specific
	PullRequestNumber int    `json:"pull_request_number,omitempty"`
	PullRequestTitle  string `json:"pull_request_title,omitempty"`
	PullRequestState  string `json:"pull_request_state,omitempty"`
	
	// Issues specific
	IssueNumber int    `json:"issue_number,omitempty"`
	IssueTitle  string `json:"issue_title,omitempty"`
	IssueState  string `json:"issue_state,omitempty"`
	
	// Installation specific
	InstallationAction string   `json:"installation_action,omitempty"`
	RepositoriesAdded  []string `json:"repositories_added,omitempty"`
	RepositoriesRemoved []string `json:"repositories_removed,omitempty"`
	
	// Raw webhook payload for advanced processing
	RawPayload map[string]interface{} `json:"raw_payload,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
}