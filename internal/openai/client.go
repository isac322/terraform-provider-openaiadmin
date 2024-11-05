// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package openai

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	Invites                InviteService
	ProjectAPIKeys         ProjectAPIKeyService
	Projects               ProjectService
	ProjectServiceAccounts ProjectServiceAccountService
	ProjectUsers           ProjectUserService
	Users                  UserService
}

func NewSDKClient(apiKey string, baseURL *string) Client {
	options := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != nil {
		options = append(options, option.WithBaseURL(*baseURL))
	}

	client := openai.NewClient(options...)
	return Client{
		Invites:                NewSDKInviteService(client),
		ProjectAPIKeys:         NewSDKProjectAPIKeyService(client),
		Projects:               NewSDKProjectService(client),
		ProjectServiceAccounts: NewSDKProjectServiceAccountService(client),
		ProjectUsers:           NewSDKProjectUserService(client),
		Users:                  NewSDKUserService(client),
	}
}
