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

type UsersListDataSource struct {
	client *openai.Client
}

type UsersListDataSourceModel struct {
	Users []UserDataSourceModel `tfsdk:"users"`
}

func NewUsersListDataSource() datasource.DataSource {
	return &UsersListDataSource{}
}

func (d *UsersListDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_users_list"
}

func (d *UsersListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve a list of all users.",

		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of all users.",
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
						"created_at": schema.StringAttribute{
							Computed: true,
						},
						"disabled": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersListDataSource) Configure(
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

func (d *UsersListDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersListDataSourceModel

	users, err := d.client.Users.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading users list", err.Error())
		return
	}

	for _, user := range users {
		data.Users = append(data.Users, UserDataSourceModel{
			ID:        types.StringValue(user.ID),
			Email:     types.StringValue(user.Email),
			Role:      types.StringValue(string(user.Role)),
			CreatedAt: types.StringValue(user.CreatedAt.Format(time.RFC3339)),
			Disabled:  types.BoolValue(user.Disabled),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
