package api

import (
	"context"
	"time"
)

type AccessRequestStatus string

const (
	StatusPending  AccessRequestStatus = "pending"
	StatusApproved AccessRequestStatus = "approved"
	StatusDenied   AccessRequestStatus = "denied"
)

type CloudResource struct {
	Type       string
	Name       string
	Project    string
	Attributes map[string]string
}

type AccessRequest struct {
	ID            string
	RequesterID   string
	RequesterName string
	Resource      CloudResource
	RequestedAt   time.Time
	Status        AccessRequestStatus
	ApproverID    string
	Command       string
	CompletedAt   time.Time
}

type AskForAccessCommand struct {
	UserID   string
	UserName string
	Message  string
}

type RespondToAccessRequestCommand struct {
	RequestID  string
	Approved   bool
	ApproverID string
	Message    string
}

type Service interface {
	AskForAccess(ctx context.Context, command AskForAccessCommand) (*AccessRequest, error)
	RespondToAccessRequest(ctx context.Context, command RespondToAccessRequestCommand) error
}