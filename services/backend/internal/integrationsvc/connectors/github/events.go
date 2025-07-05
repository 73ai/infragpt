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

type EventSubType string

const (
	EventSubTypeInstallationRepositories EventSubType = "installation_repositories"
)

type InstallationEvent struct {
	Action              string         `json:"action"`
	Installation        Installation   `json:"installation"`
	Repositories        []Repository   `json:"repositories,omitempty"`
	RepositoriesAdded   []Repository   `json:"repositories_added,omitempty"`
	RepositoriesRemoved []Repository   `json:"repositories_removed,omitempty"`
	Sender              User           `json:"sender"`
	RawPayload          map[string]any `json:"-"`
}

type Installation struct {
	ID                  int64             `json:"id"`
	AppID               int64             `json:"app_id"`
	AppSlug             string            `json:"app_slug"`
	TargetID            int64             `json:"target_id"`
	TargetType          string            `json:"target_type"`
	Account             Account           `json:"account"`
	RepositorySelection string            `json:"repository_selection"`
	AccessTokensURL     string            `json:"access_tokens_url"`
	RepositoriesURL     string            `json:"repositories_url"`
	HTMLURL             string            `json:"html_url"`
	Permissions         map[string]string `json:"permissions"`
	Events              []string          `json:"events"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	SuspendedBy         *User             `json:"suspended_by,omitempty"`
	SuspendedAt         *time.Time        `json:"suspended_at,omitempty"`
}

type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Private       bool      `json:"private"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
	DefaultBranch string    `json:"default_branch"`
}

type Account struct {
	ID      int64  `json:"id"`
	Login   string `json:"login"`
	Type    string `json:"type"`
	HTMLURL string `json:"html_url"`
}

type User struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

type WebhookEvent struct {
	EventType           EventType
	EventSubtype        EventSubType
	InstallationID      string
	RepositoryID        int64
	RepositoryName      string
	SenderID            int64
	SenderLogin         string
	Action              string
	Ref                 string
	Branch              string
	CommitSHA           string
	PullRequestNumber   int
	PullRequestTitle    string
	PullRequestState    string
	IssueNumber         int
	IssueTitle          string
	IssueState          string
	InstallationAction  string
	RepositoriesAdded   []string
	RepositoriesRemoved []string
	RawPayload          map[string]any
	CreatedAt           time.Time
}
