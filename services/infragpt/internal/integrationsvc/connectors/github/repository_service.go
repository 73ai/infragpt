package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/priyanshujain/infragpt/services/infragpt"
)

// RepositoryService manages GitHub repository permissions and tracking
type RepositoryService interface {
	// Installation management
	StoreUnclaimedInstallation(ctx context.Context, installation *UnclaimedInstallation) error
	ClaimInstallation(ctx context.Context, installationID int64, organizationID string, userID string) (*GitHubInstallation, error)
	GetUnclaimedInstallation(ctx context.Context, installationID int64) (*UnclaimedInstallation, error)
	
	// Repository tracking
	SyncRepositories(ctx context.Context, integrationID string, installationID int64) error
	AddRepositories(ctx context.Context, integrationID string, repositories []Repository) error
	RemoveRepositories(ctx context.Context, integrationID string, repositoryIDs []int64) error
	UpdateRepositoryPermissions(ctx context.Context, integrationID string, repositoryID int64, permissions RepositoryPermissions) error
	
	// Installation tracking
	CreateGitHubInstallation(ctx context.Context, integration *infragpt.Integration, installation *Installation) (*GitHubInstallation, error)
	UpdateGitHubInstallation(ctx context.Context, installationID int64, installation *Installation) error
	GetGitHubInstallation(ctx context.Context, integrationID string) (*GitHubInstallation, error)
	
	// Repository queries
	ListRepositories(ctx context.Context, integrationID string) ([]GitHubRepository, error)
	GetRepository(ctx context.Context, integrationID string, repositoryID int64) (*GitHubRepository, error)
}

// RepositoryPermissions represents repository-level permissions
type RepositoryPermissions struct {
	Admin bool
	Push  bool
	Pull  bool
}

// GitHubInstallation represents the database model for GitHub installations
type GitHubInstallation struct {
	ID                        string
	IntegrationID             string
	GitHubInstallationID      int64
	GitHubAppID               int64
	GitHubAccountID           int64
	GitHubAccountLogin        string
	GitHubAccountType         string
	RepositorySelection       string
	Permissions               map[string]string
	Events                    []string
	AppSlug                   string
	TargetType                string
	AccessTokensURL           string
	RepositoriesURL           string
	HTMLURL                   string
	SuspendedAt               *time.Time
	SuspendedBy               map[string]any
	TotalRepositoryCount      int
	AccessibleRepositoryCount int
	LastRepositorySyncAt      *time.Time
	PermissionsUpdatedAt      *time.Time
	PermissionVersion         int
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	GitHubCreatedAt           time.Time
	GitHubUpdatedAt           time.Time
}

// GitHubRepository represents the database model for GitHub repositories
type GitHubRepository struct {
	ID                     string
	IntegrationID          string
	GitHubRepositoryID     int64
	RepositoryName         string
	RepositoryFullName     string
	RepositoryURL          string
	IsPrivate              bool
	DefaultBranch          string
	PermissionAdmin        bool
	PermissionPush         bool
	PermissionPull         bool
	RepositoryDescription  string
	RepositoryLanguage     string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	LastSyncedAt           time.Time
	GitHubCreatedAt        time.Time
	GitHubUpdatedAt        time.Time
	GitHubPushedAt         time.Time
}

// repositoryServiceImpl implements the RepositoryService interface
type repositoryServiceImpl struct {
	connector *githubConnector
	// TODO: Add database repository interfaces
	// unclaimedInstallationRepo UnclaimedInstallationRepository
	// githubInstallationRepo    GitHubInstallationRepository
	// githubRepositoryRepo      GitHubRepositoryRepository
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(connector *githubConnector) RepositoryService {
	return &repositoryServiceImpl{
		connector: connector,
	}
}

func (rs *repositoryServiceImpl) StoreUnclaimedInstallation(ctx context.Context, installation *UnclaimedInstallation) error {
	slog.Info("storing unclaimed installation",
		"installation_id", installation.GitHubInstallationID,
		"account", installation.GitHubAccountLogin)

	// TODO: Implement database storage
	// return rs.unclaimedInstallationRepo.Create(ctx, installation)
	
	// For now, just log the storage
	slog.Debug("unclaimed installation stored", "installation_id", installation.GitHubInstallationID)
	return nil
}

func (rs *repositoryServiceImpl) ClaimInstallation(ctx context.Context, installationID int64, organizationID string, userID string) (*GitHubInstallation, error) {
	slog.Info("claiming installation",
		"installation_id", installationID,
		"organization_id", organizationID,
		"user_id", userID)

	// TODO: Implement claim logic
	// 1. Get unclaimed installation from database
	// 2. Create GitHubInstallation record
	// 3. Mark unclaimed installation as claimed
	// 4. Sync repositories
	
	return nil, fmt.Errorf("claim installation not yet implemented")
}

func (rs *repositoryServiceImpl) GetUnclaimedInstallation(ctx context.Context, installationID int64) (*UnclaimedInstallation, error) {
	// TODO: Implement database query
	// return rs.unclaimedInstallationRepo.GetByInstallationID(ctx, installationID)
	
	return nil, fmt.Errorf("get unclaimed installation not yet implemented")
}

func (rs *repositoryServiceImpl) SyncRepositories(ctx context.Context, integrationID string, installationID int64) error {
	slog.Info("syncing repositories",
		"integration_id", integrationID,
		"installation_id", installationID)

	// Generate JWT and get installation access token
	jwt, err := rs.connector.generateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	accessToken, err := rs.connector.getInstallationAccessToken(jwt, installationID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Fetch repositories from GitHub API
	repositories, err := rs.fetchInstallationRepositories(accessToken.Token)
	if err != nil {
		return fmt.Errorf("failed to fetch repositories: %w", err)
	}

	slog.Info("fetched repositories from GitHub",
		"integration_id", integrationID,
		"repository_count", len(repositories))

	// Store repositories in database
	for _, repo := range repositories {
		permissions := RepositoryPermissions{
			Admin: false, // TODO: Extract from API response
			Push:  false, // TODO: Extract from API response
			Pull:  true,  // Default permission for installations
		}

		if err := rs.storeRepository(ctx, integrationID, repo, permissions); err != nil {
			slog.Error("failed to store repository",
				"integration_id", integrationID,
				"repository_id", repo.ID,
				"repository_name", repo.FullName,
				"error", err)
			continue
		}
	}

	// TODO: Update last sync time
	// return rs.githubInstallationRepo.UpdateLastSyncTime(ctx, integrationID, time.Now())

	return nil
}

func (rs *repositoryServiceImpl) AddRepositories(ctx context.Context, integrationID string, repositories []Repository) error {
	slog.Info("adding repositories",
		"integration_id", integrationID,
		"repository_count", len(repositories))

	for _, repo := range repositories {
		// Default permissions for newly added repositories
		permissions := RepositoryPermissions{
			Admin: false,
			Push:  false,
			Pull:  true,
		}

		if err := rs.storeRepository(ctx, integrationID, repo, permissions); err != nil {
			slog.Error("failed to add repository",
				"integration_id", integrationID,
				"repository_id", repo.ID,
				"repository_name", repo.FullName,
				"error", err)
			continue
		}
	}

	return nil
}

func (rs *repositoryServiceImpl) RemoveRepositories(ctx context.Context, integrationID string, repositoryIDs []int64) error {
	slog.Info("removing repositories",
		"integration_id", integrationID,
		"repository_count", len(repositoryIDs))

	for _, repoID := range repositoryIDs {
		// TODO: Remove from database
		// if err := rs.githubRepositoryRepo.DeleteByGitHubID(ctx, integrationID, repoID); err != nil {
		//     slog.Error("failed to remove repository", "integration_id", integrationID, "repository_id", repoID, "error", err)
		//     continue
		// }

		slog.Debug("repository removed",
			"integration_id", integrationID,
			"repository_id", repoID)
	}

	return nil
}

func (rs *repositoryServiceImpl) UpdateRepositoryPermissions(ctx context.Context, integrationID string, repositoryID int64, permissions RepositoryPermissions) error {
	slog.Info("updating repository permissions",
		"integration_id", integrationID,
		"repository_id", repositoryID,
		"permissions", permissions)

	// TODO: Update permissions in database
	// return rs.githubRepositoryRepo.UpdatePermissions(ctx, integrationID, repositoryID, permissions)

	return nil
}

func (rs *repositoryServiceImpl) CreateGitHubInstallation(ctx context.Context, integration *infragpt.Integration, installation *Installation) (*GitHubInstallation, error) {
	githubInstallation := &GitHubInstallation{
		IntegrationID:             integration.ID,
		GitHubInstallationID:      installation.ID,
		GitHubAppID:               installation.AppID,
		GitHubAccountID:           installation.Account.ID,
		GitHubAccountLogin:        installation.Account.Login,
		GitHubAccountType:         installation.Account.Type,
		RepositorySelection:       installation.RepositorySelection,
		Permissions:               installation.Permissions,
		Events:                    installation.Events,
		AppSlug:                   installation.AppSlug,
		TargetType:                installation.TargetType,
		AccessTokensURL:           installation.AccessTokensURL,
		RepositoriesURL:           installation.RepositoriesURL,
		HTMLURL:                   installation.HTMLURL,
		SuspendedAt:               installation.SuspendedAt,
		TotalRepositoryCount:      0,
		AccessibleRepositoryCount: 0,
		LastRepositorySyncAt:      nil,
		PermissionsUpdatedAt:      &installation.UpdatedAt,
		PermissionVersion:         1,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		GitHubCreatedAt:           installation.CreatedAt,
		GitHubUpdatedAt:           installation.UpdatedAt,
	}

	// TODO: Store in database
	// return rs.githubInstallationRepo.Create(ctx, githubInstallation)

	return githubInstallation, nil
}

func (rs *repositoryServiceImpl) UpdateGitHubInstallation(ctx context.Context, installationID int64, installation *Installation) error {
	// TODO: Update installation in database
	// return rs.githubInstallationRepo.UpdateByGitHubID(ctx, installationID, installation)

	return nil
}

func (rs *repositoryServiceImpl) GetGitHubInstallation(ctx context.Context, integrationID string) (*GitHubInstallation, error) {
	// TODO: Get from database
	// return rs.githubInstallationRepo.GetByIntegrationID(ctx, integrationID)

	return nil, fmt.Errorf("get GitHub installation not yet implemented")
}

func (rs *repositoryServiceImpl) ListRepositories(ctx context.Context, integrationID string) ([]GitHubRepository, error) {
	// TODO: Get from database
	// return rs.githubRepositoryRepo.ListByIntegrationID(ctx, integrationID)

	return nil, fmt.Errorf("list repositories not yet implemented")
}

func (rs *repositoryServiceImpl) GetRepository(ctx context.Context, integrationID string, repositoryID int64) (*GitHubRepository, error) {
	// TODO: Get from database
	// return rs.githubRepositoryRepo.GetByGitHubID(ctx, integrationID, repositoryID)

	return nil, fmt.Errorf("get repository not yet implemented")
}

// Helper methods

func (rs *repositoryServiceImpl) fetchInstallationRepositories(accessToken string) ([]Repository, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/installation/repositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := rs.connector.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	var response struct {
		TotalCount   int          `json:"total_count"`
		Repositories []Repository `json:"repositories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode repositories response: %w", err)
	}

	return response.Repositories, nil
}

func (rs *repositoryServiceImpl) storeRepository(ctx context.Context, integrationID string, repo Repository, permissions RepositoryPermissions) error {
	// TODO: Store repository in database
	// githubRepo := GitHubRepository{
	// 	IntegrationID:         integrationID,
	// 	GitHubRepositoryID:    repo.ID,
	// 	RepositoryName:        repo.Name,
	// 	RepositoryFullName:    repo.FullName,
	// 	RepositoryURL:         repo.HTMLURL,
	// 	IsPrivate:             repo.Private,
	// 	DefaultBranch:         repo.DefaultBranch,
	// 	PermissionAdmin:       permissions.Admin,
	// 	PermissionPush:        permissions.Push,
	// 	PermissionPull:        permissions.Pull,
	// 	RepositoryDescription: repo.Description,
	// 	RepositoryLanguage:    repo.Language,
	// 	CreatedAt:             time.Now(),
	// 	UpdatedAt:             time.Now(),
	// 	LastSyncedAt:          time.Now(),
	// 	GitHubCreatedAt:       repo.CreatedAt,
	// 	GitHubUpdatedAt:       repo.UpdatedAt,
	// 	GitHubPushedAt:        repo.PushedAt,
	// }
	// return rs.githubRepositoryRepo.Upsert(ctx, githubRepo)

	slog.Debug("repository stored",
		"integration_id", integrationID,
		"repository_id", repo.ID,
		"repository_name", repo.FullName)

	return nil
}