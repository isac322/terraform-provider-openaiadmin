// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

// Ensure OpenAIAdminProvider satisfies various provider interfaces.
var _ provider.Provider = &OpenAIAdminProvider{}

// OpenAIAdminProvider defines the provider implementation.
type OpenAIAdminProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OpenAIAdminProviderModel describes the provider data model.
type OpenAIAdminProviderModel struct {
	BaseURL    types.String `tfsdk:"base_url"`
	AdminToken types.String `tfsdk:"admin_token"`
}

func (p *OpenAIAdminProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openaiadmin"
	resp.Version = p.version
}

func (p *OpenAIAdminProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_token": schema.StringAttribute{
				MarkdownDescription: "The Admin API key for the OpenAI API. You can create an API key at https://platform.openai.com/settings/organization/admin-keys. (Required)",
				Required:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL of the OpenAI API. (Default: `https://api.openai.com/v1`)",
				Optional:            true,
			},
		},
	}
}

func (p *OpenAIAdminProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var data OpenAIAdminProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := openai.NewClient(data.AdminToken.ValueString(), data.BaseURL.ValueStringPointer())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OpenAIAdminProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInviteResource,
		NewProjectServiceAccountResource,
		NewProjectUserResource,
		NewUserResource,
	}
}

func (p *OpenAIAdminProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInviteDataSource,
		NewInvitesByEmailDataSource,
		NewProjectAPIKeyDataSource,
		NewProjectServiceAccountDataSource,
		NewProjectUserDataSource,
		NewUserDataSource,
		NewUsersListDataSource,
		NewUserByEmailDataSource,
		NewUsersByEmailsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OpenAIAdminProvider{
			version: version,
		}
	}
}
