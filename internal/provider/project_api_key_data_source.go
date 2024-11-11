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

var _ datasource.DataSource = &ProjectAPIKeyDataSource{}

func NewProjectAPIKeyDataSource() datasource.DataSource {
	return &ProjectAPIKeyDataSource{}
}

type ProjectAPIKeyDataSource struct {
	client openai.Client
}

type ProjectAPIKeyOwnerServiceAccountModel struct {
	ID        types.String      `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	Role      types.String      `tfsdk:"role"`
}

type ProjectAPIKeyOwnerUserModel struct {
	ID        types.String      `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	Email     types.String      `tfsdk:"email"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	Role      types.String      `tfsdk:"role"`
}

type ProjectAPIKeyOwnerModel struct {
	Type           types.String                           `tfsdk:"type"`
	ServiceAccount *ProjectAPIKeyOwnerServiceAccountModel `tfsdk:"service_account"`
	User           *ProjectAPIKeyOwnerUserModel           `tfsdk:"user"`
}

type ProjectAPIKeyModel struct {
	ProjectID     types.String             `tfsdk:"project_id"`
	ID            types.String             `tfsdk:"id"`
	Name          types.String             `tfsdk:"name"`
	RedactedValue types.String             `tfsdk:"redacted_value"`
	CreatedAt     timetypes.RFC3339        `tfsdk:"created_at"`
	Owner         *ProjectAPIKeyOwnerModel `tfsdk:"owner"`
}

func (r *ProjectAPIKeyDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project_api_key"
}

func (r *ProjectAPIKeyDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Project API Key resource",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project API key.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project API key.",
				Computed:            true,
			},
			"redacted_value": schema.StringAttribute{
				MarkdownDescription: "The redacted value of the project API key.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "The timestamp when the API key was created.",
				Computed:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "The owner of the project API key.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of the owner, either 'user' or 'service_account'.",
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("user", "service_account"),
						},
					},
					"service_account": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								MarkdownDescription: "The ID of the service account.",
								Computed:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: "The name of the service account.",
								Computed:            true,
							},
							"created_at": schema.StringAttribute{
								CustomType:          timetypes.RFC3339Type{},
								MarkdownDescription: "The timestamp when the service account was created.",
								Computed:            true,
							},
							"role": schema.StringAttribute{
								MarkdownDescription: "The role of the service account.",
								Computed:            true,
								Validators: []validator.String{stringvalidator.OneOf(
									string(openai.ProjectServiceAccountRoleMember),
									string(openai.ProjectServiceAccountRoleOwner),
								)},
							},
						},
					},
					"user": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								MarkdownDescription: "The ID of the user.",
								Computed:            true,
							},
							"name": schema.StringAttribute{
								MarkdownDescription: "The name of the user.",
								Computed:            true,
								Optional:            true,
							},
							"email": schema.StringAttribute{
								MarkdownDescription: "The email of the user.",
								Computed:            true,
							},
							"created_at": schema.StringAttribute{
								CustomType:          timetypes.RFC3339Type{},
								MarkdownDescription: "The timestamp when the user was created.",
								Computed:            true,
							},
							"role": schema.StringAttribute{
								MarkdownDescription: "The role of the user.",
								Computed:            true,
								Validators: []validator.String{stringvalidator.OneOf(
									string(openai.UserRoleReader),
									string(openai.UserRoleOwner),
								)},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ProjectAPIKeyDataSource) Configure(
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
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected openai.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ProjectAPIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectAPIKeyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.ProjectAPIKeys.Retrieve(ctx, data.ProjectID.ValueString(), data.ID.ValueString())
	if err != nil {
		if openai.IsNotFoundError(err) {
			resp.Diagnostics.AddError(
				"Project API Key not found",
				fmt.Sprintf("No project API key found with ID %s.", data.ID.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading project API key", fmt.Sprintf("%+v", err))
		return
	}

	data.Name = types.StringPointerValue(apiKey.Name)
	data.RedactedValue = types.StringValue(apiKey.RedactedValue)
	data.CreatedAt = timetypes.NewRFC3339TimeValue(apiKey.CreatedAt.Time)
	data.Owner = &ProjectAPIKeyOwnerModel{}
	switch apiKey.Owner.Type {
	case "user":
		data.Owner.Type = types.StringValue("user")
		if apiKey.Owner.User == nil {
			resp.Diagnostics.AddError(
				"Error reading project API key",
				"API key owner is of type 'user' but user data is missing",
			)
			return
		}
		data.Owner.ServiceAccount = nil
		data.Owner.User = &ProjectAPIKeyOwnerUserModel{
			ID:        types.StringValue(apiKey.Owner.User.ID),
			Name:      types.StringPointerValue(apiKey.Owner.User.Name),
			Email:     types.StringValue(apiKey.Owner.User.Email),
			CreatedAt: timetypes.NewRFC3339TimeValue(apiKey.Owner.User.CreatedAt.Time),
			Role:      types.StringValue(string(apiKey.Owner.User.Role)),
		}
	case "service_account":
		data.Owner.Type = types.StringValue("service_account")
		if apiKey.Owner.ServiceAccount == nil {
			resp.Diagnostics.AddError(
				"Error reading project API key",
				"API key owner is of type 'service_account' but service account data is missing",
			)
			return
		}
		data.Owner.User = nil
		data.Owner.ServiceAccount = &ProjectAPIKeyOwnerServiceAccountModel{
			ID:        types.StringValue(apiKey.Owner.ServiceAccount.ID),
			Name:      types.StringValue(apiKey.Owner.ServiceAccount.Name),
			CreatedAt: timetypes.NewRFC3339TimeValue(apiKey.Owner.ServiceAccount.CreatedAt.Time),
			Role:      types.StringValue(string(apiKey.Owner.ServiceAccount.Role)),
		}
	}

	// Log the data source retrieval
	tflog.Trace(
		ctx,
		"Retrieved project api key",
		map[string]any{
			"id":         data.ID.ValueString(),
			"project_id": data.ProjectID.ValueString(),
		},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
