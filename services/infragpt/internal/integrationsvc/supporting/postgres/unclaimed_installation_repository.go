package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/connectors/github"
	"github.com/sqlc-dev/pqtype"
)

type unclaimedInstallationRepository struct {
	queries *Queries
}

func NewUnclaimedInstallationRepository(db *sql.DB) github.UnclaimedInstallationRepository {
	return &unclaimedInstallationRepository{queries: New(db)}
}

func (r *unclaimedInstallationRepository) Create(ctx context.Context, installation github.UnclaimedInstallation) error {
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

	err = r.queries.StoreUnclaimedInstallation(ctx, StoreUnclaimedInstallationParams{
		ID:                   installation.ID,
		GithubInstallationID: installation.GitHubInstallationID,
		GithubAppID:          installation.GitHubAppID,
		GithubAccountID:      installation.GitHubAccountID,
		GithubAccountLogin:   installation.GitHubAccountLogin,
		GithubAccountType:    installation.GitHubAccountType,
		RepositorySelection:  installation.RepositorySelection,
		Permissions:          permissionsJSON,
		Events:               installation.Events,
		AccessTokensUrl:      nullString(installation.AccessTokensURL),
		RepositoriesUrl:      nullString(installation.RepositoriesURL),
		HtmlUrl:              nullString(installation.HTMLURL),
		AppSlug:              nullString(installation.AppSlug),
		SuspendedAt:          nullTimeFromTime(installation.SuspendedAt),
		SuspendedBy:          pqtype.NullRawMessage{RawMessage: suspendedByJSON, Valid: len(suspendedByJSON) > 0},
		WebhookSender:        pqtype.NullRawMessage{RawMessage: webhookSenderJSON, Valid: len(webhookSenderJSON) > 0},
		RawWebhookPayload:    pqtype.NullRawMessage{RawMessage: rawPayloadJSON, Valid: len(rawPayloadJSON) > 0},
		CreatedAt:            installation.CreatedAt,
		GithubCreatedAt:      installation.GitHubCreatedAt,
		GithubUpdatedAt:      installation.GitHubUpdatedAt,
		ExpiresAt:            installation.ExpiresAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create unclaimed installation: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) GetByInstallationID(ctx context.Context, installationID int64) (github.UnclaimedInstallation, error) {
	dbInstallation, err := r.queries.FindUnclaimedInstallationByInstallationID(ctx, installationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return github.UnclaimedInstallation{}, nil
		}
		return github.UnclaimedInstallation{}, fmt.Errorf("failed to get unclaimed installation: %w", err)
	}

	installation := github.UnclaimedInstallation{
		ID:                     dbInstallation.ID,
		GitHubInstallationID:   dbInstallation.GithubInstallationID,
		GitHubAppID:            dbInstallation.GithubAppID,
		GitHubAccountID:        dbInstallation.GithubAccountID,
		GitHubAccountLogin:     dbInstallation.GithubAccountLogin,
		GitHubAccountType:      dbInstallation.GithubAccountType,
		RepositorySelection:    dbInstallation.RepositorySelection,
		Events:                 dbInstallation.Events,
		AccessTokensURL:        dbInstallation.AccessTokensUrl.String,
		RepositoriesURL:        dbInstallation.RepositoriesUrl.String,
		HTMLURL:                dbInstallation.HtmlUrl.String,
		AppSlug:                dbInstallation.AppSlug.String,
		SuspendedAt:            timeFromNullTime(dbInstallation.SuspendedAt),
		CreatedAt:              dbInstallation.CreatedAt,
		GitHubCreatedAt:        dbInstallation.GithubCreatedAt,
		GitHubUpdatedAt:        dbInstallation.GithubUpdatedAt,
		ExpiresAt:              dbInstallation.ExpiresAt,
		ClaimedAt:              timeFromNullTime(dbInstallation.ClaimedAt),
		ClaimedByOrganizationID: uuidFromNullUUID(dbInstallation.ClaimedByOrganizationID),
		ClaimedByUserID:        uuidFromNullUUID(dbInstallation.ClaimedByUserID),
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(dbInstallation.Permissions, &installation.Permissions); err != nil {
		return github.UnclaimedInstallation{}, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if dbInstallation.SuspendedBy.Valid && len(dbInstallation.SuspendedBy.RawMessage) > 0 {
		if err := json.Unmarshal(dbInstallation.SuspendedBy.RawMessage, &installation.SuspendedBy); err != nil {
			return github.UnclaimedInstallation{}, fmt.Errorf("failed to unmarshal suspended_by: %w", err)
		}
	}

	if dbInstallation.WebhookSender.Valid && len(dbInstallation.WebhookSender.RawMessage) > 0 {
		if err := json.Unmarshal(dbInstallation.WebhookSender.RawMessage, &installation.WebhookSender); err != nil {
			return github.UnclaimedInstallation{}, fmt.Errorf("failed to unmarshal webhook_sender: %w", err)
		}
	}

	if dbInstallation.RawWebhookPayload.Valid && len(dbInstallation.RawWebhookPayload.RawMessage) > 0 {
		if err := json.Unmarshal(dbInstallation.RawWebhookPayload.RawMessage, &installation.RawWebhookPayload); err != nil {
			return github.UnclaimedInstallation{}, fmt.Errorf("failed to unmarshal raw_webhook_payload: %w", err)
		}
	}

	return installation, nil
}

func (r *unclaimedInstallationRepository) MarkAsClaimed(ctx context.Context, installationID int64, organizationID, userID uuid.UUID) error {
	err := r.queries.MarkUnclaimedInstallationAsClaimed(ctx, MarkUnclaimedInstallationAsClaimedParams{
		ClaimedByOrganizationID: nullUUIDFromUUID(organizationID),
		ClaimedByUserID:         nullUUIDFromUUID(userID),
		GithubInstallationID:    installationID,
	})

	if err != nil {
		return fmt.Errorf("failed to mark installation as claimed: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) DeleteExpired(ctx context.Context) error {
	err := r.queries.DeleteExpiredUnclaimedInstallations(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete expired installations: %w", err)
	}

	return nil
}

func (r *unclaimedInstallationRepository) List(ctx context.Context, limit int) ([]github.UnclaimedInstallation, error) {
	dbInstallations, err := r.queries.FindUnclaimedInstallations(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to list unclaimed installations: %w", err)
	}

	var installations []github.UnclaimedInstallation

	for _, dbInstallation := range dbInstallations {
		installation := github.UnclaimedInstallation{
			ID:                     dbInstallation.ID,
			GitHubInstallationID:   dbInstallation.GithubInstallationID,
			GitHubAppID:            dbInstallation.GithubAppID,
			GitHubAccountID:        dbInstallation.GithubAccountID,
			GitHubAccountLogin:     dbInstallation.GithubAccountLogin,
			GitHubAccountType:      dbInstallation.GithubAccountType,
			RepositorySelection:    dbInstallation.RepositorySelection,
			Events:                 dbInstallation.Events,
			AccessTokensURL:        dbInstallation.AccessTokensUrl.String,
			RepositoriesURL:        dbInstallation.RepositoriesUrl.String,
			HTMLURL:                dbInstallation.HtmlUrl.String,
			AppSlug:                dbInstallation.AppSlug.String,
			SuspendedAt:            timeFromNullTime(dbInstallation.SuspendedAt),
			CreatedAt:              dbInstallation.CreatedAt,
			GitHubCreatedAt:        dbInstallation.GithubCreatedAt,
			GitHubUpdatedAt:        dbInstallation.GithubUpdatedAt,
			ExpiresAt:              dbInstallation.ExpiresAt,
			ClaimedAt:              timeFromNullTime(dbInstallation.ClaimedAt),
			ClaimedByOrganizationID: uuidFromNullUUID(dbInstallation.ClaimedByOrganizationID),
			ClaimedByUserID:        uuidFromNullUUID(dbInstallation.ClaimedByUserID),
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(dbInstallation.Permissions, &installation.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}

		if dbInstallation.SuspendedBy.Valid && len(dbInstallation.SuspendedBy.RawMessage) > 0 {
			if err := json.Unmarshal(dbInstallation.SuspendedBy.RawMessage, &installation.SuspendedBy); err != nil {
				return nil, fmt.Errorf("failed to unmarshal suspended_by: %w", err)
			}
		}

		if dbInstallation.WebhookSender.Valid && len(dbInstallation.WebhookSender.RawMessage) > 0 {
			if err := json.Unmarshal(dbInstallation.WebhookSender.RawMessage, &installation.WebhookSender); err != nil {
				return nil, fmt.Errorf("failed to unmarshal webhook_sender: %w", err)
			}
		}

		if dbInstallation.RawWebhookPayload.Valid && len(dbInstallation.RawWebhookPayload.RawMessage) > 0 {
			if err := json.Unmarshal(dbInstallation.RawWebhookPayload.RawMessage, &installation.RawWebhookPayload); err != nil {
				return nil, fmt.Errorf("failed to unmarshal raw_webhook_payload: %w", err)
			}
		}

		installations = append(installations, installation)
	}

	return installations, nil
}

func (r *unclaimedInstallationRepository) Delete(ctx context.Context, installationID int64) error {
	err := r.queries.DeleteUnclaimedInstallation(ctx, installationID)
	if err != nil {
		return fmt.Errorf("failed to delete unclaimed installation: %w", err)
	}

	return nil
}

func nullTimeFromTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func timeFromNullTime(nt sql.NullTime) time.Time {
	if !nt.Valid {
		return time.Time{}
	}
	return nt.Time
}

func nullUUIDFromUUID(u uuid.UUID) uuid.NullUUID {
	if u == uuid.Nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: u, Valid: true}
}

func uuidFromNullUUID(nu uuid.NullUUID) uuid.UUID {
	if !nu.Valid {
		return uuid.Nil
	}
	return nu.UUID
}