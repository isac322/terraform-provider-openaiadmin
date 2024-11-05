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

type ProjectUserService interface {
	List(ctx context.Context, projectID string) ([]ProjectUser, error)
	Create(ctx context.Context, projectID, userID string, role ProjectUserRole) (*ProjectUser, error)
	Retrieve(ctx context.Context, projectID, userID string) (*ProjectUser, error)
	Modify(ctx context.Context, projectID, userID string, role ProjectUserRole) (*ProjectUser, error)
	Delete(ctx context.Context, projectID, userID string) error
}

type sdkProjectUserService struct {
	client *openai.Client
}

func NewSDKProjectUserService(client *openai.Client) ProjectUserService {
	return sdkProjectUserService{client: client}
}

type ProjectUserListParams struct {
	After *string
	Limit *int
}

func (p ProjectUserListParams) URLQuery() url.Values {
	if p.After == nil && p.Limit == nil {
		return nil
	}

	v := url.Values{}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	return v
}

type ProjectUserListResponse struct {
	Data    []ProjectUser `json:"data"`
	FirstID string        `json:"first_id"`
	LastID  string        `json:"last_id"`
	HasMore bool          `json:"has_more"`
}

func (s sdkProjectUserService) List(ctx context.Context, projectID string) ([]ProjectUser, error) {
	var users []ProjectUser

	limit := 100
	params := ProjectUserListParams{
		Limit: &limit,
	}

	for {
		var result ProjectUserListResponse
		err := s.client.Get(ctx, "/organization/projects/"+projectID+"/users", params, &result)
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

type ProjectUserCreateBody struct {
	UserID string          `json:"user_id"`
	Role   ProjectUserRole `json:"role"`
}

func (s sdkProjectUserService) Create(
	ctx context.Context,
	projectID, userID string,
	role ProjectUserRole,
) (*ProjectUser, error) {
	var result ProjectUser
	err := s.client.Post(
		ctx,
		"/organization/projects/"+projectID+"/users",
		ProjectUserCreateBody{UserID: userID, Role: role},
		&result,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func (s sdkProjectUserService) Retrieve(
	ctx context.Context,
	projectID, userID string,
) (*ProjectUser, error) {
	var result ProjectUser
	err := s.client.Get(ctx, "/organization/projects/"+projectID+"/users/"+userID, nil, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

type ProjectUserModifyBody struct {
	Role ProjectUserRole `json:"role"`
}

func (s sdkProjectUserService) Modify(
	ctx context.Context,
	projectID, userID string,
	role ProjectUserRole,
) (*ProjectUser, error) {
	var result ProjectUser
	err := s.client.Post(
		ctx,
		"/organization/projects/"+projectID+"/users/"+userID,
		ProjectUserModifyBody{Role: role},
		&result,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func (s sdkProjectUserService) Delete(ctx context.Context, projectID, userID string) error {
	err := s.client.Delete(
		ctx,
		"/organization/projects/"+projectID+"/users/"+userID,
		nil,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type ProjectUserRole string

const (
	ProjectUserRoleMember ProjectUserRole = "member"
	ProjectUserRoleOwner  ProjectUserRole = "owner"
)

type ProjectUser struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Email   string              `json:"email"`
	Role    ProjectUserRole     `json:"role"`
	AddedAt utils.UnixTimestamp `json:"added_at"`
}
