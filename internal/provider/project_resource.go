// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

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
var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client openai.Client
}

type ProjectModel struct {
	ID         types.String      `tfsdk:"id"`
	Name       types.String      `tfsdk:"name"`
	Status     types.String      `tfsdk:"status"`
	CreatedAt  timetypes.RFC3339 `tfsdk:"created_at"`
	ArchivedAt timetypes.RFC3339 `tfsdk:"archived_at"`
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the project.",
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(string(openai.ProjectStatusActive), string(openai.ProjectStatusArchived)),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the project was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"archived_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the project was archived.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(
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
				"Expected openai.ProjectService, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Projects.Create(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", fmt.Sprintf("%+v", err))
		return
	}

	data.ID = types.StringValue(project.ID)
	data.Status = types.StringValue(string(project.Status))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(project.CreatedAt.Time)
	if project.ArchiveAt != nil {
		data.ArchivedAt = timetypes.NewRFC3339TimeValue(project.ArchiveAt.Time)
	} else {
		data.ArchivedAt = timetypes.NewRFC3339Null()
	}

	tflog.Trace(ctx, "Created a Project resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Projects.Retrieve(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", fmt.Sprintf("%+v", err))
		return
	}

	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(string(project.Status))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(project.CreatedAt.Time)
	if project.ArchiveAt != nil {
		data.ArchivedAt = timetypes.NewRFC3339TimeValue(project.ArchiveAt.Time)
	} else {
		data.ArchivedAt = timetypes.NewRFC3339Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.Projects.Modify(ctx, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", fmt.Sprintf("%+v", err))
		return
	}

	data.Name = types.StringValue(project.Name)
	data.Status = types.StringValue(string(project.Status))
	data.CreatedAt = timetypes.NewRFC3339TimeValue(project.CreatedAt.Time)
	if project.ArchiveAt != nil {
		data.ArchivedAt = timetypes.NewRFC3339TimeValue(project.ArchiveAt.Time)
	} else {
		data.ArchivedAt = timetypes.NewRFC3339Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Projects.Archive(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error archiving project", fmt.Sprintf("%+v", err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ProjectResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
