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

type ProjectServiceAccountResource struct {
	client *openai.Client
}

type ServiceAccountAPIKeyModel struct {
	Value     types.String `tfsdk:"value"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	ID        types.String `tfsdk:"id"`
}

type ProjectServiceAccountModel struct {
	ID        types.String               `tfsdk:"id"`
	Name      types.String               `tfsdk:"name"`
	ProjectID types.String               `tfsdk:"project_id"`
	Role      types.String               `tfsdk:"role"`
	CreatedAt types.String               `tfsdk:"created_at"`
	APIKey    *ServiceAccountAPIKeyModel `tfsdk:"api_key"` // API Key 객체로 정의
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
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.OneOf(
					string(openai.ProjectServiceAccountRoleViewer),
					string(openai.ProjectServiceAccountRoleEditor),
					string(openai.ProjectServiceAccountRoleAdmin),
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
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the API key.",
						Computed:            true,
					},
					"created_at": schema.StringAttribute{
						CustomType:          timetypes.RFC3339Type{},
						MarkdownDescription: "The timestamp when the API key was created.",
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "The ID of the API key.",
						Computed:            true,
					},
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

	if err := r.create(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error creating project service account", err.Error())
		return
	}

	tflog.Trace(ctx, "Created a Project Service Account resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectServiceAccountResource) create(ctx context.Context, data *ProjectServiceAccountModel) error {
	serviceAccount, err := r.client.ProjectServiceAccounts.Create(
		ctx,
		data.ProjectID.ValueString(),
		data.Name.ValueString(),
		openai.ProjectServiceAccountRole(data.Role.ValueString()),
	)
	if err != nil {
		return err
	}

	data.ID = types.StringValue(serviceAccount.ID)
	data.CreatedAt = types.StringValue(serviceAccount.CreatedAt.Format(time.RFC3339))
	data.Role = types.StringValue(string(serviceAccount.Role))
	data.APIKey = &ServiceAccountAPIKeyModel{
		Value:     types.StringValue(serviceAccount.APIKey.Value),
		Name:      types.StringPointerValue(serviceAccount.APIKey.Name),
		CreatedAt: types.StringValue(serviceAccount.APIKey.CreatedAt.Format(time.RFC3339)),
		ID:        types.StringValue(serviceAccount.APIKey.ID),
	}
	return nil
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
		resp.Diagnostics.AddError("Error reading project service account", err.Error())
		return
	}

	data.Name = types.StringValue(serviceAccount.Name)
	data.Role = types.StringValue(string(serviceAccount.Role))
	data.CreatedAt = types.StringValue(serviceAccount.CreatedAt.Format(time.RFC3339))
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
		resp.Diagnostics.AddError("Error deleting project service account", err.Error())
		return
	}

	if err := r.create(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error creating project service account", err.Error())
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

	if err := r.client.ProjectServiceAccounts.Delete(
		ctx,
		data.ProjectID.ValueString(),
		data.ID.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting project service account", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
