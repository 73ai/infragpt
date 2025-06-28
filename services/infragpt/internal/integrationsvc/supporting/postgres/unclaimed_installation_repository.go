package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
)

type unclaimedInstallationRepository struct {
	db *sql.DB
}

func NewUnclaimedInstallationRepository(db *sql.DB) github.UnclaimedInstallationRepository {
	return &unclaimedInstallationRepository{db: db}
}

func (r *unclaimedInstallationRepository) Create(ctx context.Context, installation *github.UnclaimedInstallation) error {
	permissionsJSON, err := json.Marshal(installation.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	suspendedByJSON, err := json.Marshal(installation.SuspendedBy)
	if err != nil {
		return fmt.Errorf("failed to marshal suspended_by: %w", err)
	}

	webhookSenderJSON, err := json.Marshal(installation.WebhookSender)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook_sender: %w", err)
	}

	rawPayloadJSON, err := json.Marshal(installation.RawWebhookPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal raw_webhook_payload: %w", err)
	}

	query := `
		INSERT INTO unclaimed_installations (
			id, github_installation_id, github_app_id, github_account_id,
			github_account_login, github_account_type, repository_selection,
			permissions, events, access_tokens_url, repositories_url, html_url,
			app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
			created_at, github_created_at, github_updated_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err = r.db.ExecContext(ctx, query,
		installation.ID,
		installation.GitHubInstallationID,
		installation.GitHubAppID,
		installation.GitHubAccountID,
		installation.GitHubAccountLogin,
		installation.GitHubAccountType,
		installation.RepositorySelection,
		permissionsJSON,
		pq.Array(installation.Events),
		nullString(installation.AccessTokensURL),
		nullString(installation.RepositoriesURL),
		nullString(installation.HTMLURL),
		nullString(installation.AppSlug),
		installation.SuspendedAt,
		suspendedByJSON,
		webhookSenderJSON,
		rawPayloadJSON,
		installation.CreatedAt,
		installation.GitHubCreatedAt,
		installation.GitHubUpdatedAt,
		installation.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create unclaimed installation: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) GetByInstallationID(ctx context.Context, installationID int64) (*github.UnclaimedInstallation, error) {
	query := `
		SELECT id, github_installation_id, github_app_id, github_account_id,
			github_account_login, github_account_type, repository_selection,
			permissions, events, access_tokens_url, repositories_url, html_url,
			app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
			created_at, github_created_at, github_updated_at, expires_at,
			claimed_at, claimed_by_organization_id, claimed_by_user_id
		FROM unclaimed_installations 
		WHERE github_installation_id = $1`

	row := r.db.QueryRowContext(ctx, query, installationID)

	installation := &github.UnclaimedInstallation{}
	var permissionsJSON, suspendedByJSON, webhookSenderJSON, rawPayloadJSON []byte
	var events pq.StringArray
	var accessTokensURL, repositoriesURL, htmlURL, appSlug sql.NullString

	err := row.Scan(
		&installation.ID,
		&installation.GitHubInstallationID,
		&installation.GitHubAppID,
		&installation.GitHubAccountID,
		&installation.GitHubAccountLogin,
		&installation.GitHubAccountType,
		&installation.RepositorySelection,
		&permissionsJSON,
		&events,
		&accessTokensURL,
		&repositoriesURL,
		&htmlURL,
		&appSlug,
		&installation.SuspendedAt,
		&suspendedByJSON,
		&webhookSenderJSON,
		&rawPayloadJSON,
		&installation.CreatedAt,
		&installation.GitHubCreatedAt,
		&installation.GitHubUpdatedAt,
		&installation.ExpiresAt,
		&installation.ClaimedAt,
		&installation.ClaimedByOrganizationID,
		&installation.ClaimedByUserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get unclaimed installation: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(permissionsJSON, &installation.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if len(suspendedByJSON) > 0 {
		if err := json.Unmarshal(suspendedByJSON, &installation.SuspendedBy); err != nil {
			return nil, fmt.Errorf("failed to unmarshal suspended_by: %w", err)
		}
	}

	if len(webhookSenderJSON) > 0 {
		if err := json.Unmarshal(webhookSenderJSON, &installation.WebhookSender); err != nil {
			return nil, fmt.Errorf("failed to unmarshal webhook_sender: %w", err)
		}
	}

	if len(rawPayloadJSON) > 0 {
		if err := json.Unmarshal(rawPayloadJSON, &installation.RawWebhookPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw_webhook_payload: %w", err)
		}
	}

	installation.Events = []string(events)
	installation.AccessTokensURL = accessTokensURL.String
	installation.RepositoriesURL = repositoriesURL.String
	installation.HTMLURL = htmlURL.String
	installation.AppSlug = appSlug.String

	return installation, nil
}

func (r *unclaimedInstallationRepository) MarkAsClaimed(ctx context.Context, installationID int64, organizationID, userID uuid.UUID) error {
	query := `
		UPDATE unclaimed_installations 
		SET claimed_at = NOW(), 
			claimed_by_organization_id = $1, 
			claimed_by_user_id = $2
		WHERE github_installation_id = $3`

	_, err := r.db.ExecContext(ctx, query, organizationID, userID, installationID)
	if err != nil {
		return fmt.Errorf("failed to mark installation as claimed: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM unclaimed_installations WHERE expires_at < NOW() AND claimed_at IS NULL`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired installations: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) List(ctx context.Context, limit int) ([]github.UnclaimedInstallation, error) {
	query := `
		SELECT id, github_installation_id, github_app_id, github_account_id,
			github_account_login, github_account_type, repository_selection,
			permissions, events, access_tokens_url, repositories_url, html_url,
			app_slug, suspended_at, suspended_by, webhook_sender, raw_webhook_payload,
			created_at, github_created_at, github_updated_at, expires_at,
			claimed_at, claimed_by_organization_id, claimed_by_user_id
		FROM unclaimed_installations 
		WHERE claimed_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list unclaimed installations: %w", err)
	}
	defer rows.Close()

	var installations []github.UnclaimedInstallation

	for rows.Next() {
		installation := github.UnclaimedInstallation{}
		var permissionsJSON, suspendedByJSON, webhookSenderJSON, rawPayloadJSON []byte
		var events pq.StringArray
		var accessTokensURL, repositoriesURL, htmlURL, appSlug sql.NullString

		err := rows.Scan(
			&installation.ID,
			&installation.GitHubInstallationID,
			&installation.GitHubAppID,
			&installation.GitHubAccountID,
			&installation.GitHubAccountLogin,
			&installation.GitHubAccountType,
			&installation.RepositorySelection,
			&permissionsJSON,
			&events,
			&accessTokensURL,
			&repositoriesURL,
			&htmlURL,
			&appSlug,
			&installation.SuspendedAt,
			&suspendedByJSON,
			&webhookSenderJSON,
			&rawPayloadJSON,
			&installation.CreatedAt,
			&installation.GitHubCreatedAt,
			&installation.GitHubUpdatedAt,
			&installation.ExpiresAt,
			&installation.ClaimedAt,
			&installation.ClaimedByOrganizationID,
			&installation.ClaimedByUserID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan unclaimed installation: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(permissionsJSON, &installation.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		if len(suspendedByJSON) > 0 {
			if err := json.Unmarshal(suspendedByJSON, &installation.SuspendedBy); err != nil {
				return nil, fmt.Errorf("failed to unmarshal suspended_by: %w", err)
			}
		}

		if len(webhookSenderJSON) > 0 {
			if err := json.Unmarshal(webhookSenderJSON, &installation.WebhookSender); err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook_sender: %w", err)
			}
		}

		if len(rawPayloadJSON) > 0 {
			if err := json.Unmarshal(rawPayloadJSON, &installation.RawWebhookPayload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal raw_webhook_payload: %w", err)
			}
		}

		installation.Events = []string(events)
		installation.AccessTokensURL = accessTokensURL.String
		installation.RepositoriesURL = repositoriesURL.String
		installation.HTMLURL = htmlURL.String
		installation.AppSlug = appSlug.String

		installations = append(installations, installation)
	}

	return installations, nil
}

func (r *unclaimedInstallationRepository) Delete(ctx context.Context, installationID int64) error {
	query := `DELETE FROM unclaimed_installations WHERE github_installation_id = $1`

	_, err := r.db.ExecContext(ctx, query, installationID)
	if err != nil {
		return fmt.Errorf("failed to delete unclaimed installation: %w", err)
	}

	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}