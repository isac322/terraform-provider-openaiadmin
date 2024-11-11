// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

type ProjectDataSource struct {
	client openai.Client
}

type ProjectDataSourceModel struct {
	ID         types.String      `tfsdk:"id"`
	Name       types.String      `tfsdk:"name"`
	Status     types.String      `tfsdk:"status"`
	CreatedAt  timetypes.RFC3339 `tfsdk:"created_at"`
	ArchivedAt timetypes.RFC3339 `tfsdk:"archived_at"`
}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

func (d *ProjectDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve details of a specific project by ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the project.",
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(openai.ProjectStatusActive), string(openai.ProjectStatusArchived)),
				},
			},
			"created_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the project was created.",
				Computed:            true,
			},
			"archived_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the project was archived.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(openai.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected openai.ProjectService, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			))
		return
	}

	d.client = client
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.Projects.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.Diagnostics.AddError(
				"Project not found",
				fmt.Sprintf("No project found with ID %s.", data.ID.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading project", fmt.Sprintf("%+v", err))
		return
	}

	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(string(project.Status))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(project.CreatedAt.Time)
	if project.ArchiveAt != nil {
		data.ArchivedAt = timetypes.NewRFC3339TimeValue(project.ArchiveAt.Time)
	} else {
		data.ArchivedAt = timetypes.NewRFC3339Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
