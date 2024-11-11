// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/stretchr/testify/require"
)

func TestAccProjectServiceAccountDataSource(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	serviceAccount, err := client.ProjectServiceAccounts.Create(
		ctx,
		project.ID,
		"test-acc-ds-sa",
	)
	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		require.NoError(t, client.ProjectServiceAccounts.Delete(ctx, project.ID, serviceAccount.ID))
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectServiceAccountDataSourceConfig(
					project.ID,
					serviceAccount.ID,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.openaiadmin_project_service_account.test",
						"id",
						serviceAccount.ID,
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_project_service_account.test",
						"name",
						serviceAccount.Name,
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_project_service_account.test",
						"project_id",
						project.ID,
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_project_service_account.test",
						"role",
						string(serviceAccount.Role),
					),
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_project_service_account.test",
						"created_at",
					),
				),
			},
		},
	})
}

func TestAccProjectServiceAccountDataSource_NonExistent(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	projectID := "test-project-id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectServiceAccountDataSourceConfig(
					projectID,
					"non-existent-id",
				),
				ExpectError: regexp.MustCompile(
					`Service Account not found`,
				),
			},
		},
	})
}

func testAccProjectServiceAccountDataSourceConfig(projectID, saID string) string {
	return fmt.Sprintf(`
data "openaiadmin_project_service_account" "test" {
  id         = %[1]q
  project_id = %[2]q
}
`, saID, projectID)
}
