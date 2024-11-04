// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

type InviteDataSource struct {
	client *openai.Client
}

type InviteDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	InvitedAt  types.String `tfsdk:"invited_at"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	AcceptedAt types.String `tfsdk:"accepted_at"`
}

func NewInviteDataSource() datasource.DataSource {
	return &InviteDataSource{}
}

func (d *InviteDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (d *InviteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve details of a specific invite by ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invite.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email associated with the invite.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the invite.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the invite.",
				Computed:            true,
			},
			"invited_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite was created.",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite expires.",
				Computed:            true,
			},
			"accepted_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite was accepted.",
				Computed:            true,
			},
		},
	}
}

func (d *InviteDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*openai.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			))
		return
	}

	d.client = client
}

func (d *InviteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InviteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite, err := d.client.Invites.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading invite", err.Error())
		return
	}

	data.Email = types.StringValue(invite.Email)
	data.Role = types.StringValue(string(invite.Role))
	data.Status = types.StringValue(string(invite.Status))
	data.InvitedAt = types.StringValue(invite.InvitedAt.Format(time.RFC3339))
	data.ExpiresAt = types.StringValue(invite.ExpiresAt.Format(time.RFC3339))
	if invite.AcceptedAt != nil {
		data.AcceptedAt = types.StringValue(invite.AcceptedAt.Format(time.RFC3339))
	} else {
		data.AcceptedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
