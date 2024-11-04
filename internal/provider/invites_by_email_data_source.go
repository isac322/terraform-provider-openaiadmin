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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

type InvitesByEmailDataSource struct {
	client *openai.Client
}

type InviteData struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	InvitedAt  types.String `tfsdk:"invited_at"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	AcceptedAt types.String `tfsdk:"accepted_at"`
}

func NewInvitesByEmailDataSource() datasource.DataSource {
	return &InvitesByEmailDataSource{}
}

func (d *InvitesByEmailDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_invites_by_email"
}

func (d *InvitesByEmailDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve a list of invitations for a given email and optional status.",

		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "The email to filter invitations.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Optional status to filter invitations.",
				Optional:            true,
			},
			"invites": schema.ListNestedAttribute{
				MarkdownDescription: "List of invitations matching the criteria.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"role": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"invited_at": schema.StringAttribute{
							Computed: true,
						},
						"expires_at": schema.StringAttribute{
							Computed: true,
						},
						"accepted_at": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *InvitesByEmailDataSource) Configure(
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

func (d *InvitesByEmailDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var data struct {
		Email   types.String `tfsdk:"email"`
		Status  types.String `tfsdk:"status"`
		Invites []InviteData `tfsdk:"invites"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allInvites, err := d.client.Invites.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading invites list", err.Error())
		return
	}

	for _, invite := range allInvites {
		if invite.Email == data.Email.ValueString() && (data.Status.IsNull() || string(invite.Status) == data.Status.ValueString()) {
			data.Invites = append(data.Invites, InviteData{
				ID:        types.StringValue(invite.ID),
				Email:     types.StringValue(invite.Email),
				Role:      types.StringValue(string(invite.Role)),
				Status:    types.StringValue(string(invite.Status)),
				InvitedAt: types.StringValue(invite.InvitedAt.Format(time.RFC3339)),
				ExpiresAt: types.StringValue(invite.ExpiresAt.Format(time.RFC3339)),
				AcceptedAt: func() types.String {
					if invite.AcceptedAt != nil {
						return types.StringValue(invite.AcceptedAt.Format(time.RFC3339))
					}
					return types.StringNull()
				}(),
			})
		}
	}

	tflog.Trace(ctx, "Retrieved invites by email", map[string]interface{}{
		"email":      data.Email.ValueString(),
		"status":     data.Status,
		"invite_ids": data.Invites,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
