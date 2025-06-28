package github

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UnclaimedInstallationRepository interface {
	Create(ctx context.Context, installation UnclaimedInstallation) error
	GetByInstallationID(ctx context.Context, installationID int64) (UnclaimedInstallation, error)
	MarkAsClaimed(ctx context.Context, installationID int64, organizationID, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	List(ctx context.Context, limit int) ([]UnclaimedInstallation, error)
	Delete(ctx context.Context, installationID int64) error
}

type GitHubRepositoryRepository interface {
	Upsert(ctx context.Context, repo *GitHubRepository) error
	ListByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]GitHubRepository, error)
	GetByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) (GitHubRepository, error)
	DeleteByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) error
	UpdatePermissions(ctx context.Context, integrationID uuid.UUID, repositoryID int64, permissions RepositoryPermissions) error
	BulkDelete(ctx context.Context, integrationID uuid.UUID, repositoryIDs []int64) error
	UpdateLastSyncTime(ctx context.Context, integrationID uuid.UUID, syncTime time.Time) error
}

type UnclaimedInstallation struct {
	ID                      uuid.UUID
	GitHubInstallationID    int64
	GitHubAppID             int64
	GitHubAccountID         int64
	GitHubAccountLogin      string
	GitHubAccountType       string
	RepositorySelection     string
	Permissions             map[string]string
	Events                  []string
	AccessTokensURL         string
	RepositoriesURL         string
	HTMLURL                 string
	AppSlug                 string
	SuspendedAt             time.Time
	SuspendedBy             map[string]any
	WebhookSender           map[string]any
	RawWebhookPayload       map[string]any
	CreatedAt               time.Time
	GitHubCreatedAt         time.Time
	GitHubUpdatedAt         time.Time
	ExpiresAt               time.Time
	ClaimedAt               time.Time
	ClaimedByOrganizationID uuid.UUID
	ClaimedByUserID         uuid.UUID
}

type GitHubRepository struct {
	ID                    uuid.UUID
	IntegrationID         uuid.UUID
	GitHubRepositoryID    int64
	RepositoryName        string
	RepositoryFullName    string
	RepositoryURL         string
	IsPrivate             bool
	DefaultBranch         string
	PermissionAdmin       bool
	PermissionPush        bool
	PermissionPull        bool
	RepositoryDescription string
	RepositoryLanguage    string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	LastSyncedAt          time.Time
	GitHubCreatedAt       time.Time
	GitHubUpdatedAt       time.Time
	GitHubPushedAt        time.Time
}

type RepositoryPermissions struct {
	Admin bool
	Push  bool
	Pull  bool
}