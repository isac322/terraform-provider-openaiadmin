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

type ProjectServiceAccountService interface {
	List(ctx context.Context, projectID string) ([]ProjectServiceAccount, error)
	Create(ctx context.Context, projectID, name string) (*ProjectServiceAccountWithAPIKey, error)
	Retrieve(ctx context.Context, projectID, serviceAccountID string) (*ProjectServiceAccount, error)
	Delete(ctx context.Context, projectID, serviceAccountID string) error
}

// sdkProjectServiceAccountService handles operations related to project service accounts in the OpenAI admin API.
type sdkProjectServiceAccountService struct {
	client *openai.Client
}

func NewSDKProjectServiceAccountService(client *openai.Client) ProjectServiceAccountService {
	return sdkProjectServiceAccountService{client: client}
}

// ProjectServiceAccountRole represents the possible roles of a project service account.
type ProjectServiceAccountRole string

const (
	ProjectServiceAccountRoleMember ProjectServiceAccountRole = "member"
	ProjectServiceAccountRoleOwner  ProjectServiceAccountRole = "owner"
	ProjectServiceAccountRoleAdmin  ProjectServiceAccountRole = "admin"
)

type ProjectServiceAccount struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	ProjectID string                    `json:"project_id"`
	CreatedAt utils.UnixTimestamp       `json:"created_at"`
	Role      ProjectServiceAccountRole `json:"role"`
}

type ProjectServiceAccountListParams struct {
	After *string
	Limit *int
}

func (p ProjectServiceAccountListParams) URLQuery() url.Values {
	v := url.Values{}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	return v
}

type ProjectServiceAccountListResponse struct {
	Data    []ProjectServiceAccount `json:"data"`
	HasMore bool                    `json:"has_more"`
	LastID  string                  `json:"last_id"`
}

// List retrieves all service accounts for a project, with optional pagination parameters.
func (s sdkProjectServiceAccountService) List(ctx context.Context, projectID string) ([]ProjectServiceAccount, error) {
	var accounts []ProjectServiceAccount

	limit := 100
	params := ProjectServiceAccountListParams{
		Limit: &limit,
	}

	for {
		var result ProjectServiceAccountListResponse
		err := s.client.Get(ctx, "/organization/projects/"+projectID+"/service_accounts", params, &result)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		accounts = append(accounts, result.Data...)
		if !result.HasMore {
			break
		}
		params.After = &result.LastID
	}

	return accounts, nil
}

type ProjectServiceAccountCreateBody struct {
	Name string `json:"name"`
}

type ServiceAccountAPIKey struct {
	Value     string              `json:"value"`
	Name      *string             `json:"name"`
	CreatedAt utils.UnixTimestamp `json:"created_at"`
	ID        string              `json:"id"`
}

type ProjectServiceAccountWithAPIKey struct {
	ProjectServiceAccount
	APIKey ServiceAccountAPIKey `json:"api_key"`
}

// Create adds a new service account to the project.
func (s sdkProjectServiceAccountService) Create(
	ctx context.Context,
	projectID, name string,
) (*ProjectServiceAccountWithAPIKey, error) {
	var result ProjectServiceAccountWithAPIKey
	body := ProjectServiceAccountCreateBody{Name: name}
	err := s.client.Post(ctx, "/organization/projects/"+projectID+"/service_accounts", body, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Retrieve fetches details of a project service account by its ID.
func (s sdkProjectServiceAccountService) Retrieve(
	ctx context.Context,
	projectID, serviceAccountID string,
) (*ProjectServiceAccount, error) {
	var result ProjectServiceAccount
	err := s.client.Get(
		ctx,
		"/organization/projects/"+projectID+"/service_accounts/"+serviceAccountID,
		nil,
		&result,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Delete removes a service account from the project by its ID.
func (s sdkProjectServiceAccountService) Delete(ctx context.Context, projectID, serviceAccountID string) error {
	err := s.client.Delete(
		ctx,
		"/organization/projects/"+projectID+"/service_accounts/"+serviceAccountID,
		nil,
		nil,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
