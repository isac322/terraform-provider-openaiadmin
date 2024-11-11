// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package openai

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pkg/errors"
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
	options := []option.RequestOption{option.WithAPIKey(apiKey), option.WithMaxRetries(5)}
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

func IsNotFoundError(err error) bool {
	var openaiErr *openai.Error
	if errors.As(err, &openaiErr) {
		return openaiErr.StatusCode == 404
	}
	return false
}
