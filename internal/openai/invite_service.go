// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package openai

import (
	"context"
	"net/url"
	"strconv"

	"github.com/isac322/terraform-provider-openaiadmin/internal/utils"
	"github.com/openai/openai-go"
	"github.com/pkg/errors"
)

// InviteService handles operations related to invites in the OpenAI admin API.
type InviteService struct {
	client *openai.Client
}

func NewInviteService(client *openai.Client) *InviteService {
	return &InviteService{client: client}
}

// InviteStatus represents the possible statuses of an invite.
type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusExpired  InviteStatus = "expired"
)

// InviteRole represents the possible roles of an invite.
type InviteRole string

const (
	InviteRoleMember InviteRole = "member"
	InviteRoleAdmin  InviteRole = "admin"
)

type Invite struct {
	ID         string               `json:"id"`
	Email      string               `json:"email"`
	Status     InviteStatus         `json:"status"`
	Role       InviteRole           `json:"role"`
	InvitedAt  utils.UnixTimestamp  `json:"invited_at"`
	ExpiresAt  utils.UnixTimestamp  `json:"expires_at,omitempty"`
	AcceptedAt *utils.UnixTimestamp `json:"accepted_at,omitempty"`
}

type InviteListParams struct {
	After *string
	Limit *int
}

func (p InviteListParams) URLQuery() url.Values {
	v := url.Values{}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	return v
}

type InviteListResponse struct {
	Data    []Invite `json:"data"`
	HasMore bool     `json:"has_more"`
	LastID  string   `json:"last_id"`
}

// List retrieves all invites, with optional pagination parameters.
func (s *InviteService) List(ctx context.Context) ([]Invite, error) {
	var invites []Invite

	limit := 100
	params := InviteListParams{
		Limit: &limit,
	}

	for {
		var result InviteListResponse
		err := s.client.Get(ctx, "/organization/invites", params, &result)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		invites = append(invites, result.Data...)
		if !result.HasMore {
			break
		}
		params.After = &result.LastID
	}

	return invites, nil
}

type InviteCreateBody struct {
	Email string     `json:"email"`
	Role  InviteRole `json:"role"`
}

// Create sends an invite to a new user.
func (s *InviteService) Create(ctx context.Context, email string, role InviteRole) (*Invite, error) {
	var result Invite
	err := s.client.Post(ctx, "/organization/invites", InviteCreateBody{Email: email, Role: role}, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Retrieve fetches an invite by ID.
func (s *InviteService) Retrieve(ctx context.Context, inviteID string) (*Invite, error) {
	var result Invite
	err := s.client.Get(ctx, "/organization/invites/"+inviteID, nil, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Delete removes an invite by ID.
func (s *InviteService) Delete(ctx context.Context, inviteID string) error {
	err := s.client.Delete(ctx, "/organization/invites/"+inviteID, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
