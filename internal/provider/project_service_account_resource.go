// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectServiceAccountResource{}

type ProjectServiceAccountResource struct {
	client openai.Client
}

type ServiceAccountAPIKeyModel struct {
	Value     types.String      `tfsdk:"value"`
	Name      types.String      `tfsdk:"name"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	ID        types.String      `tfsdk:"id"`
}

type ProjectServiceAccountModel struct {
	ID        types.String      `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	ProjectID types.String      `tfsdk:"project_id"`
	Role      types.String      `tfsdk:"role"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	APIKey    types.Object      `tfsdk:"api_key"`
}

func NewProjectServiceAccountResource() resource.Resource {
	return &ProjectServiceAccountResource{}
}

func (r *ProjectServiceAccountResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project_service_account"
}

func (r *ProjectServiceAccountResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project Service Account resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project service account.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project service account.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project to which this service account belongs.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"api_key": schema.SingleNestedAttribute{
				MarkdownDescription: "The API key for the service account, available only during creation.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"value": schema.StringAttribute{
						MarkdownDescription: "The actual API key value.",
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the API key.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"created_at": schema.StringAttribute{
						CustomType:          timetypes.RFC3339Type{},
						MarkdownDescription: "The timestamp when the API key was created.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "The ID of the API key.",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectServiceAccountResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
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

func (r *ProjectServiceAccountResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data ProjectServiceAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.create(ctx, &data, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Created a Project Service Account resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) create(
	ctx context.Context,
	data *ProjectServiceAccountModel,
	diagnostics diag.Diagnostics,
) {
	serviceAccount, err := r.client.ProjectServiceAccounts.Create(
		ctx,
		data.ProjectID.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		diagnostics.AddError("Error creating project service account", fmt.Sprintf("%+v", err))
		return
	}

	data.ID = types.StringValue(serviceAccount.ID)
	data.CreatedAt = timetypes.NewRFC3339TimeValue(serviceAccount.CreatedAt.Time)
	data.Role = types.StringValue(string(serviceAccount.Role))
	var diags diag.Diagnostics
	data.APIKey, diags = types.ObjectValue(
		map[string]attr.Type{
			"value":      types.StringType,
			"name":       types.StringType,
			"created_at": timetypes.RFC3339Type{},
			"id":         types.StringType,
		},
		map[string]attr.Value{
			"value":      types.StringValue(serviceAccount.APIKey.Value),
			"name":       types.StringPointerValue(serviceAccount.APIKey.Name),
			"created_at": timetypes.NewRFC3339TimeValue(serviceAccount.APIKey.CreatedAt.Time),
			"id":         types.StringValue(serviceAccount.APIKey.ID),
		},
	)
	diagnostics.Append(diags...)
}

func (r *ProjectServiceAccountResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data ProjectServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccount, err := r.client.ProjectServiceAccounts.Retrieve(
		ctx,
		data.ProjectID.ValueString(),
		data.ID.ValueString(),
	)
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project service account", fmt.Sprintf("%+v", err))
		return
	}

	data.Name = types.StringValue(serviceAccount.Name)
	data.Role = types.StringValue(string(serviceAccount.Role))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(serviceAccount.CreatedAt.Time)
	// The API Key is not provided when searching, so do not change it.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data ProjectServiceAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.ProjectServiceAccounts.Delete(ctx, data.ProjectID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting project service account", fmt.Sprintf("%+v", err))
		return
	}

	r.create(ctx, &data, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data ProjectServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.ProjectServiceAccounts.Delete(
		ctx,
		data.ProjectID.ValueString(),
		data.ID.ValueString(),
	)
	if err != nil && !openai.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Error deleting project service account", fmt.Sprintf("%+v", err))
		return
	}

	resp.State.RemoveResource(ctx)
}
