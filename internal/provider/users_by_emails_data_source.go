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

type UsersByEmailsDataSource struct {
	client *openai.Client
}

type UserData struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.String `tfsdk:"created_at"`
	Disabled  types.Bool   `tfsdk:"disabled"`
}

func NewUsersByEmailsDataSource() datasource.DataSource {
	return &UsersByEmailsDataSource{}
}

func (d *UsersByEmailsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_users_by_emails"
}

func (d *UsersByEmailsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve details of users by an array of emails.",

		Attributes: map[string]schema.Attribute{
			"emails": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Array of user emails to search for.",
				Required:            true,
			},
			"users": schema.MapNestedAttribute{
				MarkdownDescription: "Map of emails to user details.",
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

func (d *UsersByEmailsDataSource) Configure(
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

func (d *UsersByEmailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data struct {
		Emails []types.String      `tfsdk:"emails"`
		Users  map[string]UserData `tfsdk:"users"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert emails list to a map for faster lookup and track found users
	emailsToFind := make(map[string]struct{})
	for _, email := range data.Emails {
		emailsToFind[email.ValueString()] = struct{}{}
	}

	usersFound := make(map[string]UserData)
	missingEmails := []string{}

	// Retrieve the full list of users
	allUsers, err := d.client.Users.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading users list", err.Error())
		return
	}

	// Match users by email
	for _, user := range allUsers {
		if _, ok := emailsToFind[user.Email]; ok {
			usersFound[user.Email] = UserData{
				ID:        types.StringValue(user.ID),
				Email:     types.StringValue(user.Email),
				Role:      types.StringValue(string(user.Role)),
				CreatedAt: types.StringValue(user.CreatedAt.Format(time.RFC3339)),
				Disabled:  types.BoolValue(user.Disabled),
			}
			delete(emailsToFind, user.Email) // Remove found email from search map
		}
	}

	// Any remaining emails in emailsToFind are missing
	for missingEmail := range emailsToFind {
		missingEmails = append(missingEmails, missingEmail)
	}

	// Log a warning for missing emails
	if len(missingEmails) > 0 {
		resp.Diagnostics.AddWarning(
			"Users Not Found",
			fmt.Sprintf("The following emails were not found: %v", missingEmails),
		)
	}

	// Assign found users to state
	data.Users = usersFound

	tflog.Trace(ctx, "Retrieved users by emails", map[string]interface{}{
		"emails_requested": data.Emails,
		"emails_found":     usersFound,
		"emails_missing":   missingEmails,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
