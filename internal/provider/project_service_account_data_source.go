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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// ProjectServiceAccountDataSource is the data source implementation.
type ProjectServiceAccountDataSource struct {
	client openai.Client
}

type ProjectServiceAccountDataSourceModel struct {
	ID        types.String      `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	ProjectID types.String      `tfsdk:"project_id"`
	Role      types.String      `tfsdk:"role"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
}

func NewProjectServiceAccountDataSource() datasource.DataSource {
	return &ProjectServiceAccountDataSource{}
}

func (d *ProjectServiceAccountDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project_service_account"
}

func (d *ProjectServiceAccountDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for retrieving details of a project service account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project service account.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project service account.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project to which this service account belongs.",
				Required:            true,
			},
			"created_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the service account was created.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the project service account.",
				Computed:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					string(openai.ProjectServiceAccountRoleMember),
					string(openai.ProjectServiceAccountRoleOwner),
				)},
			},
		},
	}
}

func (d *ProjectServiceAccountDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(openai.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	d.client = client
}

func (d *ProjectServiceAccountDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data ProjectServiceAccountDataSourceModel

	// Read configuration from state
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve service account information
	serviceAccount, err := d.client.ProjectServiceAccounts.Retrieve(
		ctx,
		data.ProjectID.ValueString(),
		data.ID.ValueString(),
	)
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.Diagnostics.AddError(
				"Service Account not found",
				fmt.Sprintf("No service account found with ID %s.", data.ID.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading project service account", fmt.Sprintf("%+v", err))
		return
	}

	// Populate the data source model with retrieved information
	data.Name = types.StringValue(serviceAccount.Name)
	data.Role = types.StringValue(string(serviceAccount.Role))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(serviceAccount.CreatedAt.Time)

	// Log the data source retrieval
	tflog.Trace(
		ctx,
		"Retrieved project service account",
		map[string]any{
			"id":         data.ID.ValueString(),
			"project_id": data.ProjectID.ValueString(),
		},
	)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
