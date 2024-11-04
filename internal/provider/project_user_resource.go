// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

type ProjectUserResource struct {
	client *openai.Client
}

type ProjectUserModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Email     types.String `tfsdk:"email"`
	ProjectID types.String `tfsdk:"project_id"`
	UserID    types.String `tfsdk:"user_id"`
	Role      types.String `tfsdk:"role"`
	AddedAt   types.String `tfsdk:"added_at"`
}

func NewProjectUserResource() resource.Resource {
	return &ProjectUserResource{}
}

func (r *ProjectUserResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project_user"
}

func (r *ProjectUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project User resource",

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user to be added to the project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role of the project user.",
				Required:            true,
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

func (r *ProjectUserResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*openai.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *openai.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

func (r *ProjectUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectUserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create project user
	projectUser, err := r.client.ProjectUsers.Create(
		ctx,
		data.ProjectID.ValueString(),
		data.UserID.ValueString(),
		openai.ProjectUserRole(data.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project user", err.Error())
		return
	}

	data.ID = types.StringValue(projectUser.ID)
	data.Name = types.StringValue(projectUser.Name)
	data.Email = types.StringValue(projectUser.Email)
	data.Role = types.StringValue(string(projectUser.Role))
	data.AddedAt = types.StringValue(projectUser.AddedAt.Format(time.RFC3339))

	tflog.Trace(ctx, "Created a Project User resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectUserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve project user details
	projectUser, err := r.client.ProjectUsers.Retrieve(ctx, data.ProjectID.ValueString(), data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project user", err.Error())
		return
	}

	data.Name = types.StringValue(projectUser.Name)
	data.Email = types.StringValue(projectUser.Email)
	data.Role = types.StringValue(string(projectUser.Role))
	data.AddedAt = types.StringValue(projectUser.AddedAt.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectUserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update project user role
	projectUser, err := r.client.ProjectUsers.Modify(
		ctx,
		data.ProjectID.ValueString(),
		data.UserID.ValueString(),
		openai.ProjectUserRole(data.Role.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project user", err.Error())
		return
	}

	data.Role = types.StringValue(string(projectUser.Role))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectUserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ProjectUsers.Delete(ctx, data.ProjectID.ValueString(), data.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting project user", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
