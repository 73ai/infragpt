package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
)

type githubRepositoryRepository struct {
	db *sql.DB
}

func NewGitHubRepositoryRepository(db *sql.DB) github.GitHubRepositoryRepository {
	return &githubRepositoryRepository{db: db}
}

func (r *githubRepositoryRepository) Upsert(ctx context.Context, repo *github.GitHubRepository) error {
	query := `
		INSERT INTO github_repositories (
			id, integration_id, github_repository_id, repository_name,
			repository_full_name, repository_url, is_private, default_branch,
			permission_admin, permission_push, permission_pull,
			repository_description, repository_language, created_at, updated_at,
			last_synced_at, github_created_at, github_updated_at, github_pushed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (integration_id, github_repository_id)
		DO UPDATE SET
			repository_name = EXCLUDED.repository_name,
			repository_full_name = EXCLUDED.repository_full_name,
			repository_url = EXCLUDED.repository_url,
			is_private = EXCLUDED.is_private,
			default_branch = EXCLUDED.default_branch,
			permission_admin = EXCLUDED.permission_admin,
			permission_push = EXCLUDED.permission_push,
			permission_pull = EXCLUDED.permission_pull,
			repository_description = EXCLUDED.repository_description,
			repository_language = EXCLUDED.repository_language,
			updated_at = EXCLUDED.updated_at,
			last_synced_at = EXCLUDED.last_synced_at,
			github_updated_at = EXCLUDED.github_updated_at,
			github_pushed_at = EXCLUDED.github_pushed_at`

	_, err := r.db.ExecContext(ctx, query,
		repo.ID,
		repo.IntegrationID,
		repo.GitHubRepositoryID,
		repo.RepositoryName,
		repo.RepositoryFullName,
		repo.RepositoryURL,
		repo.IsPrivate,
		nullString(repo.DefaultBranch),
		repo.PermissionAdmin,
		repo.PermissionPush,
		repo.PermissionPull,
		nullString(repo.RepositoryDescription),
		nullString(repo.RepositoryLanguage),
		repo.CreatedAt,
		repo.UpdatedAt,
		repo.LastSyncedAt,
		nullTime(repo.GitHubCreatedAt),
		nullTime(repo.GitHubUpdatedAt),
		nullTime(repo.GitHubPushedAt),
	)

	if err != nil {
		return fmt.Errorf("failed to upsert github repository: %w", err)
	}

	return nil
}

func (r *githubRepositoryRepository) ListByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]github.GitHubRepository, error) {
	query := `
		SELECT id, integration_id, github_repository_id, repository_name,
			repository_full_name, repository_url, is_private, default_branch,
			permission_admin, permission_push, permission_pull,
			repository_description, repository_language, created_at, updated_at,
			last_synced_at, github_created_at, github_updated_at, github_pushed_at
		FROM github_repositories 
		WHERE integration_id = $1
		ORDER BY repository_full_name`

	rows, err := r.db.QueryContext(ctx, query, integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list github repositories: %w", err)
	}
	defer rows.Close()

	var repositories []github.GitHubRepository

	for rows.Next() {
		repo := github.GitHubRepository{}
		var defaultBranch, description, language sql.NullString
		var githubCreatedAt, githubUpdatedAt, githubPushedAt sql.NullTime

		err := rows.Scan(
			&repo.ID,
			&repo.IntegrationID,
			&repo.GitHubRepositoryID,
			&repo.RepositoryName,
			&repo.RepositoryFullName,
			&repo.RepositoryURL,
			&repo.IsPrivate,
			&defaultBranch,
			&repo.PermissionAdmin,
			&repo.PermissionPush,
			&repo.PermissionPull,
			&description,
			&language,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&repo.LastSyncedAt,
			&githubCreatedAt,
			&githubUpdatedAt,
			&githubPushedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan github repository: %w", err)
		}

		repo.DefaultBranch = defaultBranch.String
		repo.RepositoryDescription = description.String
		repo.RepositoryLanguage = language.String
		repo.GitHubCreatedAt = githubCreatedAt.Time
		repo.GitHubUpdatedAt = githubUpdatedAt.Time
		repo.GitHubPushedAt = githubPushedAt.Time

		repositories = append(repositories, repo)
	}

	return repositories, nil
}

func (r *githubRepositoryRepository) GetByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) (*github.GitHubRepository, error) {
	query := `
		SELECT id, integration_id, github_repository_id, repository_name,
			repository_full_name, repository_url, is_private, default_branch,
			permission_admin, permission_push, permission_pull,
			repository_description, repository_language, created_at, updated_at,
			last_synced_at, github_created_at, github_updated_at, github_pushed_at
		FROM github_repositories 
		WHERE integration_id = $1 AND github_repository_id = $2`

	row := r.db.QueryRowContext(ctx, query, integrationID, repositoryID)

	repo := &github.GitHubRepository{}
	var defaultBranch, description, language sql.NullString
	var githubCreatedAt, githubUpdatedAt, githubPushedAt sql.NullTime

	err := row.Scan(
		&repo.ID,
		&repo.IntegrationID,
		&repo.GitHubRepositoryID,
		&repo.RepositoryName,
		&repo.RepositoryFullName,
		&repo.RepositoryURL,
		&repo.IsPrivate,
		&defaultBranch,
		&repo.PermissionAdmin,
		&repo.PermissionPush,
		&repo.PermissionPull,
		&description,
		&language,
		&repo.CreatedAt,
		&repo.UpdatedAt,
		&repo.LastSyncedAt,
		&githubCreatedAt,
		&githubUpdatedAt,
		&githubPushedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get github repository: %w", err)
	}

	repo.DefaultBranch = defaultBranch.String
	repo.RepositoryDescription = description.String
	repo.RepositoryLanguage = language.String
	repo.GitHubCreatedAt = githubCreatedAt.Time
	repo.GitHubUpdatedAt = githubUpdatedAt.Time
	repo.GitHubPushedAt = githubPushedAt.Time

	return repo, nil
}

func (r *githubRepositoryRepository) DeleteByGitHubID(ctx context.Context, integrationID uuid.UUID, repositoryID int64) error {
	query := `DELETE FROM github_repositories WHERE integration_id = $1 AND github_repository_id = $2`

	_, err := r.db.ExecContext(ctx, query, integrationID, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete github repository: %w", err)
	}

	return nil
}

func (r *githubRepositoryRepository) UpdatePermissions(ctx context.Context, integrationID uuid.UUID, repositoryID int64, permissions github.RepositoryPermissions) error {
	query := `
		UPDATE github_repositories 
		SET permission_admin = $1, 
			permission_push = $2, 
			permission_pull = $3,
			updated_at = NOW()
		WHERE integration_id = $4 AND github_repository_id = $5`

	_, err := r.db.ExecContext(ctx, query,
		permissions.Admin,
		permissions.Push,
		permissions.Pull,
		integrationID,
		repositoryID,
	)

	if err != nil {
		return fmt.Errorf("failed to update repository permissions: %w", err)
	}

	return nil
}

func (r *githubRepositoryRepository) BulkDelete(ctx context.Context, integrationID uuid.UUID, repositoryIDs []int64) error {
	if len(repositoryIDs) == 0 {
		return nil
	}

	query := `DELETE FROM github_repositories WHERE integration_id = $1 AND github_repository_id = ANY($2)`

	_, err := r.db.ExecContext(ctx, query, integrationID, repositoryIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk delete github repositories: %w", err)
	}

	return nil
}

func (r *githubRepositoryRepository) UpdateLastSyncTime(ctx context.Context, integrationID uuid.UUID, syncTime time.Time) error {
	query := `
		UPDATE github_repositories 
		SET last_synced_at = $1, updated_at = NOW()
		WHERE integration_id = $2`

	_, err := r.db.ExecContext(ctx, query, syncTime, integrationID)
	if err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	return nil
}

func nullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}