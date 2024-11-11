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

func TestAccProjectUserDataSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	require.NoError(t, err)

	userID := os.Getenv("OPENAI_TEST_USER_ID")
	resourceName := "data.openaiadmin_project_user.test"

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	// Create a project user for testing the data source
	projectUser, err := client.ProjectUsers.Create(
		ctx,
		project.ID,
		userID,
		openai.ProjectUserRoleMember,
	)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccProjectUserDataSourceConfig(
					project.ID,
					userID,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify computed values
					resource.TestCheckResourceAttr(resourceName, "project_id", project.ID),
					resource.TestCheckResourceAttr(resourceName, "user_id", userID),
					resource.TestCheckResourceAttr(resourceName, "name", projectUser.Name),
					resource.TestCheckResourceAttr(resourceName, "email", projectUser.Email),
					resource.TestCheckResourceAttr(resourceName, "role", string(projectUser.Role)),
					resource.TestCheckResourceAttrSet(resourceName, "added_at"),
				),
			},
		},
	})
}

func TestAccProjectUserDataSource_NonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	require.NoError(t, err)

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Try to read a non-existent user
				Config: testAccProjectUserDataSourceConfig(
					project.ID,
					"non-existent-user-id",
				),
				ExpectError: regexp.MustCompile(`Error reading project user`),
			},
		},
	})
}

// Generate test config for the data source
func testAccProjectUserDataSourceConfig(projectID, userID string) string {
	return fmt.Sprintf(`
data "openaiadmin_project_user" "test" {
  project_id = %[1]q
  user_id    = %[2]q
}
`, projectID, userID)
}
