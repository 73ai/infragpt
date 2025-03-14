package postgres

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/domain"
	db2 "github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/postgres/db"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/supporting/token"
	"time"

	"github.com/google/uuid"
	"github.com/priyanshujain/infragpt/services/infragpt/identity"
)

// TODO: load timezones

type sessionRepository struct {
	db         *sql.DB
	queries    *db2.Queries
	privateKey *rsa.PrivateKey
}

func (s sessionRepository) StartUserSession(ctx context.Context, session identity.UserSession) (
	identity.Credentials, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return identity.Credentials{}, err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	uid, _ := uuid.Parse(session.UserID)
	sid, _ := uuid.Parse(session.SessionID)

	did := newDeviceID()
	err = qtx.CreateDevice(ctx, db2.CreateDeviceParams{
		DeviceID:          did,
		UserID:            uid,
		DeviceFingerprint: session.Device.DeviceFingerprint,
		Name:              session.Device.Name,
		Os:                string(session.Device.OS),
		Brand:             session.Device.Brand,
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create device: %w", err)
	}

	err = qtx.CreateSession(ctx, db2.CreateSessionParams{
		SessionID:    sid,
		UserID:       uid,
		DeviceID:     did,
		IpAddress:    session.IPAddress,
		IpCountryIso: session.IPCountryISO,
		UserAgent:    session.UserAgent,
		Timezone:     session.Timezone.String(),
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create session: %w", err)
	}

	creds, err := s.createRefreshToken(ctx, qtx, uid, sid)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create refresh token: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("commit transaction: %w", err)
	}

	return creds, nil
}

func (s sessionRepository) createRefreshToken(ctx context.Context, qtx *db2.Queries, uid, sid uuid.UUID) (identity.Credentials, error) {
	// create refesh token
	tokenManager := token.NewManager(s.privateKey)
	refreshToken, err := tokenManager.NewRefreshToken(sid.String())
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create refresh token: %w", err)
	}

	tid, _ := uuid.Parse(refreshToken.TokenID)
	err = qtx.CreateRefreshToken(ctx, db2.CreateRefreshTokenParams{
		TokenID:   tid,
		SessionID: sid,
		UserID:    uid,
		TokenHash: refreshToken.HashedToken,
		ExpiryAt:  refreshToken.ExpiryAt,
	})
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create refresh token: %w", err)
	}

	// create access token
	accessToken, err := tokenManager.NewAccessToken(sid.String())
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create access token: %w", err)
	}

	return identity.Credentials{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.TokenString,
	}, nil
}

func (s sessionRepository) RefreshToken(ctx context.Context, tokenID string) (identity.Credentials, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)
	tid, _ := uuid.Parse(tokenID)

	refreshToken, err := qtx.RefreshToken(ctx, tid)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("get refresh token: %w", err)
	}

	err = qtx.RevokeRefreshToken(ctx, tid)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("revoke refresh token: %w", err)
	}

	creds, err := s.createRefreshToken(ctx, qtx, refreshToken.UserID, refreshToken.SessionID)
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("create refresh token: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return identity.Credentials{}, fmt.Errorf("commit transaction: %w", err)
	}

	return creds, nil
}

func (s sessionRepository) UserSessions(ctx context.Context, userID string) ([]identity.UserSession, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)
	uid, _ := uuid.Parse(userID)
	sessions, err := qtx.UserSessions(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user sessions: %w", err)
	}

	devices, err := qtx.DevicesByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get devices by user id: %w", err)
	}

	var userDevices = make(map[uuid.UUID]identity.UserDevice)
	for _, device := range devices {
		userDevices[device.DeviceID] = identity.UserDevice{
			UserID:            device.UserID.String(),
			DeviceFingerprint: device.DeviceFingerprint,
			OS:                identity.OperatingSystem(device.Os),
			Name:              device.Name,
			Brand:             device.Brand,
		}
	}

	var userSessions []identity.UserSession
	for _, session := range sessions {
		// load timezone
		loc, err := time.LoadLocation(session.Timezone)
		if err != nil {
			return nil, fmt.Errorf("load timezone: %w", err)
		}
		userSessions = append(userSessions, identity.UserSession{
			UserID:         session.UserID.String(),
			SessionID:      session.SessionID.String(),
			IPAddress:      session.IpAddress,
			UserAgent:      session.UserAgent,
			IPCountryISO:   session.IpCountryIso,
			LastActivityAt: session.LastActivityAt,
			Timezone:       loc,
			Device:         userDevices[session.DeviceID],
		})
	}

	return userSessions, nil
}

func (s sessionRepository) UserSession(ctx context.Context, sessionID string) (identity.UserSession, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return identity.UserSession{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)
	sid, _ := uuid.Parse(sessionID)
	session, err := qtx.UserSession(ctx, sid)
	if err != nil {
		return identity.UserSession{}, fmt.Errorf("get user session: %w", err)
	}

	device, err := qtx.Device(ctx, session.DeviceID)
	if err != nil {
		return identity.UserSession{}, fmt.Errorf("get device by id: %w", err)
	}

	// load timezone
	loc, err := time.LoadLocation(session.Timezone)
	if err != nil {
		return identity.UserSession{}, fmt.Errorf("load timezone: %w", err)
	}

	return identity.UserSession{
		UserID:         session.UserID.String(),
		SessionID:      session.SessionID.String(),
		IPAddress:      session.IpAddress,
		UserAgent:      session.UserAgent,
		IPCountryISO:   session.IpCountryIso,
		LastActivityAt: session.LastActivityAt,
		Timezone:       loc,
		Device: identity.UserDevice{
			UserID:            device.UserID.String(),
			DeviceFingerprint: device.DeviceFingerprint,
			OS:                identity.OperatingSystem(device.Os),
			Name:              device.Name,
			Brand:             device.Brand,
		},
	}, nil
}

func NewSessionRepository(sqlDB *sql.DB, privateKey *rsa.PrivateKey) domain.SessionRepository {
	return &sessionRepository{
		db:         sqlDB,
		queries:    db2.New(sqlDB),
		privateKey: privateKey,
	}
}

func newDeviceID() uuid.UUID {
	return uuid.New()
}
