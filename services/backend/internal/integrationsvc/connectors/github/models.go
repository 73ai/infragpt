package github

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type GitHubRepositoryRepository interface {
	Store(ctx context.Context, repo GitHubRepository) error
	ListByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]GitHubRepository, error)
	GetByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) (GitHubRepository, error)
	DeleteByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) error
	UpdatePermissions(ctx context.Context, integrationID uuid.UUID, repositoryID int64, permissions RepositoryPermissions) error
	BulkDelete(ctx context.Context, integrationID uuid.UUID, repositoryIDs []int64) error
	UpdateLastSyncTime(ctx context.Context, integrationID uuid.UUID, syncTime time.Time) error
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
