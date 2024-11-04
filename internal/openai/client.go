// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package openai

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	client *openai.Client

	Invites                *InviteService
	ProjectAPIKeys         *ProjectAPIKeyService
	ProjectServiceAccounts *ProjectServiceAccountService
	ProjectUsers           *ProjectUserService
	Users                  *UserService
}

func NewClient(apiKey string, baseURL *string) *Client {
	options := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		options = append(options, option.WithBaseURL(*baseURL))
	}

	client := openai.NewClient(options...)
	return &Client{
		client:                 client,
		Invites:                NewInviteService(client),
		ProjectAPIKeys:         NewProjectAPIKeyService(client),
		ProjectServiceAccounts: NewProjectServiceAccountService(client),
		ProjectUsers:           NewProjectUserService(client),
		Users:                  NewUserService(client),
	}
}
