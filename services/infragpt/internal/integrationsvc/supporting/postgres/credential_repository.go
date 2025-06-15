package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/integrationsvc/domain"
)

type credentialRepository struct {
	queries    *Queries
	encryption *encryptionService
}

func NewCredentialRepository(sqlDB *sql.DB) (domain.CredentialRepository, error) {
	encryption, err := newEncryptionService()
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption service: %w", err)
	}

	return &credentialRepository{
		queries:    New(sqlDB),
		encryption: encryption,
	}, nil
}

func (r *credentialRepository) Store(ctx context.Context, cred domain.IntegrationCredential) error {
	encryptedData, err := r.encryption.encrypt(cred.Data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential data: %w", err)
	}

	credentialID, err := uuid.Parse(cred.ID)
	if err != nil {
		return fmt.Errorf("invalid credential ID: %w", err)
	}

	integrationID, err := uuid.Parse(cred.IntegrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	var expiresAt sql.NullTime
	if cred.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *cred.ExpiresAt, Valid: true}
	}

	return r.queries.StoreCredential(ctx, StoreCredentialParams{
		ID:                      credentialID,
		IntegrationID:           integrationID,
		CredentialType:          string(cred.CredentialType),
		CredentialDataEncrypted: encryptedData,
		ExpiresAt:               expiresAt,
		EncryptionKeyID:         cred.EncryptionKeyID,
		CreatedAt:               cred.CreatedAt,
		UpdatedAt:               cred.UpdatedAt,
	})
}

func (r *credentialRepository) FindByIntegration(ctx context.Context, integrationID string) (domain.IntegrationCredential, error) {
	integrationUUID, err := uuid.Parse(integrationID)
	if err != nil {
		return domain.IntegrationCredential{}, fmt.Errorf("invalid integration ID: %w", err)
	}

	dbCredential, err := r.queries.FindCredentialByIntegration(ctx, integrationUUID)
	if err != nil {
		return domain.IntegrationCredential{}, fmt.Errorf("failed to find credential: %w", err)
	}

	return r.mapToCredential(dbCredential)
}

func (r *credentialRepository) Update(ctx context.Context, cred domain.IntegrationCredential) error {
	encryptedData, err := r.encryption.encrypt(cred.Data)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential data: %w", err)
	}

	integrationID, err := uuid.Parse(cred.IntegrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	var expiresAt sql.NullTime
	if cred.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *cred.ExpiresAt, Valid: true}
	}

	return r.queries.UpdateCredential(ctx, UpdateCredentialParams{
		IntegrationID:           integrationID,
		CredentialType:          string(cred.CredentialType),
		CredentialDataEncrypted: encryptedData,
		ExpiresAt:               expiresAt,
		EncryptionKeyID:         cred.EncryptionKeyID,
	})
}

func (r *credentialRepository) Delete(ctx context.Context, integrationID string) error {
	integrationUUID, err := uuid.Parse(integrationID)
	if err != nil {
		return fmt.Errorf("invalid integration ID: %w", err)
	}

	return r.queries.DeleteCredential(ctx, integrationUUID)
}

func (r *credentialRepository) FindExpiring(ctx context.Context, before time.Time) ([]domain.IntegrationCredential, error) {
	dbCredentials, err := r.queries.FindExpiringCredentials(ctx, sql.NullTime{Time: before, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to find expiring credentials: %w", err)
	}

	credentials := make([]domain.IntegrationCredential, len(dbCredentials))
	for i, dbCredential := range dbCredentials {
		credential, err := r.mapToCredential(dbCredential)
		if err != nil {
			return nil, fmt.Errorf("failed to map credential: %w", err)
		}
		credentials[i] = credential
	}

	return credentials, nil
}

func (r *credentialRepository) mapToCredential(dbCredential IntegrationCredential) (domain.IntegrationCredential, error) {
	decryptedData, err := r.encryption.decrypt(dbCredential.CredentialDataEncrypted)
	if err != nil {
		return domain.IntegrationCredential{}, fmt.Errorf("failed to decrypt credential data: %w", err)
	}

	var expiresAt *time.Time
	if dbCredential.ExpiresAt.Valid {
		expiresAt = &dbCredential.ExpiresAt.Time
	}

	return domain.IntegrationCredential{
		ID:              dbCredential.ID.String(),
		IntegrationID:   dbCredential.IntegrationID.String(),
		CredentialType:  infragpt.CredentialType(dbCredential.CredentialType),
		Data:            decryptedData,
		ExpiresAt:       expiresAt,
		EncryptionKeyID: dbCredential.EncryptionKeyID,
		CreatedAt:       dbCredential.CreatedAt,
		UpdatedAt:       dbCredential.UpdatedAt,
	}, nil
}
