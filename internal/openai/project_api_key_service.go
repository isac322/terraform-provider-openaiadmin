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

// ProjectAPIKeyService handles operations related to project API keys in the OpenAI admin API.
type ProjectAPIKeyService struct {
	client *openai.Client
}

func NewProjectAPIKeyService(client *openai.Client) *ProjectAPIKeyService {
	return &ProjectAPIKeyService{client: client}
}

type ProjectAPIKeyOwner struct {
	Type           string `json:"type"`
	ServiceAccount *struct {
		ID        string                    `json:"id"`
		Name      string                    `json:"name"`
		CreatedAt utils.UnixTimestamp       `json:"created_at"`
		Role      ProjectServiceAccountRole `json:"role"`
	} `json:"service_account,omitempty"`
	User *struct {
		ID        string              `json:"id"`
		Name      *string             `json:"name"`
		Email     string              `json:"email"`
		CreatedAt utils.UnixTimestamp `json:"created_at"`
		Role      UserRole            `json:"role"`
	} `json:"user,omitempty"`
}

type ProjectAPIKey struct {
	ID            string              `json:"id"`
	Name          *string             `json:"name"`
	RedactedValue string              `json:"redacted_value"`
	CreatedAt     utils.UnixTimestamp `json:"created_at"`
	Owner         ProjectAPIKeyOwner  `json:"owner"`
}

type ProjectAPIKeyListParams struct {
	After *string
	Limit *int
}

func (p ProjectAPIKeyListParams) URLQuery() url.Values {
	v := url.Values{}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	return v
}

type ProjectAPIKeyListResponse struct {
	Data    []ProjectAPIKey `json:"data"`
	HasMore bool            `json:"has_more"`
	LastID  string          `json:"last_id"`
}

// List retrieves all API keys for a project, with optional pagination parameters.
func (s *ProjectAPIKeyService) List(ctx context.Context, projectID string) ([]ProjectAPIKey, error) {
	var apiKeys []ProjectAPIKey

	limit := 100
	params := ProjectAPIKeyListParams{
		Limit: &limit,
	}

	for {
		var result ProjectAPIKeyListResponse
		err := s.client.Get(ctx, "/organization/projects/"+projectID+"/api-keys", params.URLQuery(), &result)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		apiKeys = append(apiKeys, result.Data...)
		if !result.HasMore {
			break
		}
		params.After = &result.LastID
	}

	return apiKeys, nil
}

// Retrieve fetches details of a project API key by its ID.
func (s *ProjectAPIKeyService) Retrieve(ctx context.Context, projectID, apiKeyID string) (*ProjectAPIKey, error) {
	var result ProjectAPIKey
	err := s.client.Get(ctx, "/organization/projects/"+projectID+"/api-keys/"+apiKeyID, nil, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

// Delete removes an API key from the project by its ID.
func (s *ProjectAPIKeyService) Delete(ctx context.Context, projectID, apiKeyID string) error {
	err := s.client.Delete(ctx, "/organization/projects/"+projectID+"/api-keys/"+apiKeyID, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
