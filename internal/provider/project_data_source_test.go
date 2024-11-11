// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/stretchr/testify/require"
)

func TestAccProjectDataSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	// Ensure cleanup of project
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectDataSourceConfig(project.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_project.test", "id", project.ID),
					resource.TestCheckResourceAttr("data.openaiadmin_project.test", "name", projectName),
					resource.TestCheckResourceAttr("data.openaiadmin_project.test", "status", string(openai.ProjectStatusActive)),
				),
			},
		},
	})
}

func testAccProjectDataSourceConfig(projectID string) string {
	return fmt.Sprintf(`
data "openaiadmin_project" "test" {
  id = "%s"
}
`, projectID)
}
