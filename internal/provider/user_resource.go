// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

type UserResource struct {
	client openai.Client
}

type UserModel struct {
	ID       types.String      `tfsdk:"id"`
	Email    types.String      `tfsdk:"email"`
	Role     types.String      `tfsdk:"role"`
	AddedAt  timetypes.RFC3339 `tfsdk:"added_at"`
	Disabled types.Bool        `tfsdk:"disabled"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "User resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the user.",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the user.",
				Computed:            true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(openai.UserRoleReader), string(openai.UserRoleOwner)),
				},
				//PlanModifiers: []planmodifier.String{
				//	stringplanmodifier.UseStateForUnknown(),
				//},
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

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(openai.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

func (r *UserResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Cannot Create User",
		"User creation is not supported by this resource. Please import an existing user instead.",
	)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Users.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", fmt.Sprintf("%+v", err))
		return
	}

	data.Email = types.StringValue(user.Email)
	data.Role = types.StringValue(string(user.Role))
	data.AddedAt = timetypes.NewRFC3339TimeValue(user.AddedAt.Time)
	data.Disabled = types.BoolValue(user.Disabled)

	tflog.Trace(ctx, "Retrieved user", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if ID is being changed
	if !data.ID.Equal(state.ID) {
		resp.Diagnostics.AddError(
			"Cannot update user ID",
			"User ID cannot be changed. Please remove the resource and import the new user instead",
		)
		return
	}

	user, err := r.client.Users.Modify(ctx, data.ID.ValueString(), openai.UserRole(data.Role.ValueString()))
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.Diagnostics.AddError(
				"Cannot update User",
				fmt.Sprintf(
					"The user %s was not found. It may have been deleted outside of Terraform.",
					data.ID.ValueString(),
				),
			)
			return
		}
		resp.Diagnostics.AddError("Error updating user", fmt.Sprintf("%+v", err))
		return
	}

	data.Email = types.StringValue(user.Email)
	data.Role = types.StringValue(string(user.Role))
	data.AddedAt = timetypes.NewRFC3339TimeValue(user.AddedAt.Time)
	data.Disabled = types.BoolValue(user.Disabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Users.Delete(ctx, data.ID.ValueString()); err != nil && !openai.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Error deleting user", fmt.Sprintf("%+v", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *UserResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
