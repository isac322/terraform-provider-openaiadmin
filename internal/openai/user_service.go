// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate mockgen -package "$GOPACKAGE" -destination "mock_$GOFILE" -source "$GOFILE" -typed

package openai

import (
	"context"
	"net/url"
	"strconv"

	"github.com/isac322/terraform-provider-openaiadmin/internal/utils"
	"github.com/openai/openai-go"
	"github.com/pkg/errors"
)

type UserService interface {
	List(ctx context.Context) ([]User, error)
	Modify(ctx context.Context, userID string, role UserRole, disabled *bool) (*User, error)
	Retrieve(ctx context.Context, userID string) (*User, error)
	Delete(ctx context.Context, userID string) error
}

// sdkUserService handles operations related to users in the OpenAI admin API.
type sdkUserService struct {
	client *openai.Client
}

func NewSDKUserService(client *openai.Client) UserService {
	return sdkUserService{client: client}
}

// UserRole represents the possible roles of a user.
type UserRole string

const (
	UserRoleMember UserRole = "member"
	UserRoleAdmin  UserRole = "admin"
)

// User represents a user in the OpenAI system.
type User struct {
	ID        string              `json:"id"`
	Email     string              `json:"email"`
	Role      UserRole            `json:"role"`
	CreatedAt utils.UnixTimestamp `json:"created_at"`
	Disabled  bool                `json:"disabled"`
}

type UserListParams struct {
	After *string
	Limit *int
}

func (p UserListParams) URLQuery() url.Values {
	v := url.Values{}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	return v
}

type UserListResponse struct {
	Data    []User `json:"data"`
	HasMore bool   `json:"has_more"`
	LastID  string `json:"last_id"`
}

// List retrieves all users, with optional pagination parameters.
func (s sdkUserService) List(ctx context.Context) ([]User, error) {
	var users []User

	limit := 100
	params := UserListParams{
		Limit: &limit,
	}

	for {
		var result UserListResponse
		err := s.client.Get(ctx, "/organization/users", params, &result)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		users = append(users, result.Data...)
		if !result.HasMore {
			break
		}
		params.After = &result.LastID
	}

	return users, nil
}

type UserModifyBody struct {
	Role     UserRole `json:"role,omitempty"`
	Disabled *bool    `json:"disabled,omitempty"`
}

// Modify updates a user's role or disabled status.
func (s sdkUserService) Modify(ctx context.Context, userID string, role UserRole, disabled *bool) (*User, error) {
	var result User
	body := UserModifyBody{Role: role, Disabled: disabled}
	err := s.client.Post(ctx, "/organization/users/"+userID, body, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Retrieve fetches a user by ID.
func (s sdkUserService) Retrieve(ctx context.Context, userID string) (*User, error) {
	var result User
	err := s.client.Get(ctx, "/organization/users/"+userID, nil, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Delete removes a user by ID.
func (s sdkUserService) Delete(ctx context.Context, userID string) error {
	err := s.client.Delete(ctx, "/organization/users/"+userID, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
