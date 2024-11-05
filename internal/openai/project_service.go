// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package openai

//go:generate mockgen -package "$GOPACKAGE" -destination "mock_$GOFILE" -source "$GOFILE" -typed

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/openai/openai-go"
	"github.com/pkg/errors"
)

type ProjectService interface {
	List(ctx context.Context) ([]Project, error)
	Create(ctx context.Context, name string) (*Project, error)
	Retrieve(ctx context.Context, projectID string) (*Project, error)
	Modify(ctx context.Context, projectID, name string) (*Project, error)
	Archive(ctx context.Context, projectID string) error
}

// SDKProjectService handles operations related to projects in the OpenAI admin API.
type SDKProjectService struct {
	client *openai.Client
}

func NewSDKProjectService(client *openai.Client) ProjectService {
	return SDKProjectService{client: client}
}

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

// Project represents a project in the OpenAI system.
type Project struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	ArchiveAt time.Time     `json:"archive_at"`
	Status    ProjectStatus `json:"status"`
}

// ProjectListParams represents the query parameters for listing projects.
type ProjectListParams struct {
	Limit *int
	After *string
}

func (p ProjectListParams) URLQuery() url.Values {
	v := url.Values{}
	if p.Limit != nil {
		v.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.After != nil {
		v.Set("after", *p.After)
	}
	return v
}

// List retrieves a list of projects.
func (s SDKProjectService) List(ctx context.Context) ([]Project, error) {
	var projects []Project

	limit := 100
	params := ProjectListParams{
		Limit: &limit,
	}

	for {
		var result struct {
			Data    []Project `json:"data"`
			HasMore bool      `json:"has_more"`
			LastID  string    `json:"last_id"`
		}

		err := s.client.Get(ctx, "/organization/projects", params.URLQuery(), &result)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		projects = append(projects, result.Data...)
		if !result.HasMore {
			break
		}
		params.After = &result.LastID
	}

	return projects, nil
}

// ProjectCreateParams represents the parameters for creating a project.
type ProjectCreateParams struct {
	Name string `json:"name"`
}

// Create creates a new project with the given parameters.
func (s SDKProjectService) Create(ctx context.Context, name string) (*Project, error) {
	var project Project
	err := s.client.Post(ctx, "/organization/projects", ProjectCreateParams{Name: name}, &project)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &project, nil
}

// Retrieve fetches a project by its ID.
func (s SDKProjectService) Retrieve(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	err := s.client.Get(ctx, "/organization/projects/"+projectID, nil, &project)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &project, nil
}

// ProjectModifyParams represents the parameters for modifying a project.
type ProjectModifyParams struct {
	Name string `json:"name"`
}

// Modify updates a project's details with the given parameters.
func (s SDKProjectService) Modify(ctx context.Context, projectID, name string) (*Project, error) {
	var project Project
	err := s.client.Post(ctx, "/organization/projects/"+projectID, ProjectModifyParams{Name: name}, &project)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &project, nil
}

// Archive archives a project by its ID.
func (s SDKProjectService) Archive(ctx context.Context, projectID string) error {
	err := s.client.Post(ctx, "/organization/projects/"+projectID+"/archive", nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
