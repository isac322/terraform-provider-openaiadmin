// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

type UserDataSource struct {
	client openai.Client
}

type UserDataSourceModel struct {
	ID       types.String      `tfsdk:"id"`
	Email    types.String      `tfsdk:"email"`
	Role     types.String      `tfsdk:"role"`
	AddedAt  timetypes.RFC3339 `tfsdk:"added_at"`
	Disabled types.Bool        `tfsdk:"disabled"`
}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

func (d *UserDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve details of a specific user by ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the user.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the user.",
				Computed:            true,
			},
			"added_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the user was created.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is disabled.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(
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
				"Expected openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			))
		return
	}

	d.client = client
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.Users.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", fmt.Sprintf("%+v", err))
		return
	}

	data.Email = types.StringValue(user.Email)
	data.Role = types.StringValue(string(user.Role))
	data.AddedAt = timetypes.NewRFC3339TimeValue(user.AddedAt.Time)
	data.Disabled = types.BoolValue(user.Disabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
