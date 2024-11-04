// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InviteResource{}
var _ resource.ResourceWithImportState = &InviteResource{}

func NewInviteResource() resource.Resource {
	return &InviteResource{}
}

type InviteResource struct {
	client *openai.Client
}

type InviteModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Role       types.String `tfsdk:"role"`
	Status     types.String `tfsdk:"status"`
	InvitedAt  types.String `tfsdk:"invited_at"`
	ExpiresAt  types.String `tfsdk:"expires_at"`
	AcceptedAt types.String `tfsdk:"accepted_at"`
}

func (r *InviteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invite"
}

func (r *InviteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Invite resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invite.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email to invite.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the invite.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(openai.InviteRoleMember), string(openai.InviteRoleAdmin)),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the invite.",
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(openai.InviteStatusPending),
						string(openai.InviteStatusAccepted),
						string(openai.InviteStatusExpired),
					),
				},
			},
			"invited_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite was created.",
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite expires.",
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
			},
			"accepted_at": schema.StringAttribute{
				MarkdownDescription: "The time the invite was accepted.",
				CustomType:          timetypes.RFC3339Type{},
				Computed:            true,
			},
		},
	}
}

func (r *InviteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*openai.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *InviteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InviteModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.createInvite(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error creating invite", err.Error())
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) createInvite(ctx context.Context, data *InviteModel) error {
	// Convert role to InviteRole type
	var role openai.InviteRole
	switch data.Role.ValueString() {
	case string(openai.InviteRoleMember):
		role = openai.InviteRoleMember
	case string(openai.InviteRoleAdmin):
		role = openai.InviteRoleAdmin
	default:
		return fmt.Errorf("role %s is not valid", data.Role.ValueString())
	}

	invite, err := r.client.Invites.Create(ctx, data.Email.ValueString(), role)
	if err != nil {
		return err
	}

	data.ID = types.StringValue(invite.ID)
	data.Status = types.StringValue(string(invite.Status))
	data.InvitedAt = types.StringValue(invite.InvitedAt.Format(time.RFC3339))
	data.ExpiresAt = types.StringValue(invite.ExpiresAt.Format(time.RFC3339))
	if invite.AcceptedAt != nil {
		data.AcceptedAt = types.StringValue(invite.AcceptedAt.Format(time.RFC3339))
	} else {
		data.AcceptedAt = types.StringNull()
	}

	return nil
}

func (r *InviteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InviteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite, err := r.client.Invites.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading invite", err.Error())
		return
	}

	data.Email = types.StringValue(invite.Email)
	data.Status = types.StringValue(string(invite.Status))
	data.Role = types.StringValue(string(invite.Role))
	data.InvitedAt = types.StringValue(invite.InvitedAt.Format(time.RFC3339))
	data.ExpiresAt = types.StringValue(invite.ExpiresAt.Format(time.RFC3339))
	if invite.AcceptedAt != nil {
		data.AcceptedAt = types.StringValue(invite.AcceptedAt.Format(time.RFC3339))
	} else {
		data.AcceptedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InviteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Invites.Delete(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting invite", err.Error())
		return
	}

	if err := r.createInvite(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error creating invite", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InviteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InviteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Invites.Delete(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting invite", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *InviteResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
