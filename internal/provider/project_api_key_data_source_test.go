// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/stretchr/testify/require"
)

func TestAccProjectAPIKeyDataSource(t *testing.T) {
	t.Parallel()

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	ctx := context.Background()
	projectName := generateTestProject()

	// Pre-test: Create an API Key
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	sa, err := client.ProjectServiceAccounts.Create(
		ctx,
		project.ID,
		generateTestServiceAccount(),
	)
	if err != nil {
		require.NoError(t, err)
	}

	// Ensure cleanup of API Key
	t.Cleanup(func() {
		require.NoError(t, client.ProjectServiceAccounts.Delete(ctx, project.ID, sa.ID))
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectAPIKeyDataSourceConfig(project.ID, sa.APIKey.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_project_api_key.test", "id", sa.APIKey.ID),
					resource.TestCheckResourceAttr("data.openaiadmin_project_api_key.test", "project_id", project.ID),
					resource.TestMatchResourceAttr(
						"data.openaiadmin_project_api_key.test",
						"redacted_value",
						regexp.MustCompile(`^sk-svcac`), // Redacted format check
					),
				),
			},
		},
	})
}

func testAccProjectAPIKeyDataSourceConfig(projectID, apiKeyID string) string {
	return fmt.Sprintf(`
data "openaiadmin_project_api_key" "test" {
  project_id = "%s"
  id         = "%s"
}
`, projectID, apiKeyID)
}

func generateTestProject() string {
	return fmt.Sprintf("test_project_%d", time.Now().UnixNano())
}

func generateTestServiceAccount() string {
	return fmt.Sprintf("test_sa_%d", time.Now().UnixNano())
}
