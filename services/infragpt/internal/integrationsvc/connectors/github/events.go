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
	EventType EventType
	InstallationID int64
	RepositoryID   int64
	RepositoryName string
	SenderID       int64
	SenderLogin    string
	Action string
	Ref      string
	Branch   string
	CommitSHA string
	PullRequestNumber int
	PullRequestTitle  string
	PullRequestState  string
	IssueNumber int
	IssueTitle  string
	IssueState  string
	InstallationAction string
	RepositoriesAdded  []string
	RepositoriesRemoved []string
	RawPayload map[string]any
	CreatedAt time.Time
}