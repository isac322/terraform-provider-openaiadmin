// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// ProjectUserDataSource is the data source implementation.
type ProjectUserDataSource struct {
	client *openai.Client
}

type ProjectUserDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Email     types.String `tfsdk:"email"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.String `tfsdk:"added_at"`
}

func NewProjectUserDataSource() datasource.DataSource {
	return &ProjectUserDataSource{}
}

func (d *ProjectUserDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}

func (d *ProjectUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for retrieving details of a project user.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project user.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the project user.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project to which this user belongs.",
				Required:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user to be added to the project.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the project user.",
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(openai.ProjectUserRoleMember), string(openai.ProjectUserRoleOwner)),
				},
			},
			"added_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the user was added to the project.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectUserDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*openai.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

func (d *ProjectUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectUserDataSourceModel

	// Read configuration from state
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve project user information
	projectUser, err := d.client.ProjectUsers.Retrieve(ctx, data.ProjectID.ValueString(), data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project user", err.Error())
		return
	}

	// Populate the data source model with retrieved information
	data.ID = types.StringValue(projectUser.ID)
	data.Name = types.StringValue(projectUser.Name)
	data.Email = types.StringValue(projectUser.Email)
	data.Role = types.StringValue(string(projectUser.Role))
	data.AddedAt = types.StringValue(projectUser.AddedAt.Format(time.RFC3339))

	// Log the data source retrieval
	tflog.Trace(ctx, "Retrieved project user", map[string]interface{}{
		"id":         data.ID.ValueString(),
		"project_id": data.ProjectID.ValueString(),
		"user_id":    data.UserID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
