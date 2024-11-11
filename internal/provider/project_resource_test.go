// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAccProjectResource_basic(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Initial project name for creation
	projectName := generateTestProject()
	updatedProjectName := generateTestProject()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(client),
		Steps: []resource.TestStep{
			// Step 1: Create Project
			{
				Config: testAccProjectResourceConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_project.test", "name", projectName),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "id"),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "status"),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "created_at"),
				),
			},
			// Step 2: Update Project
			{
				Config: testAccProjectResourceConfig(updatedProjectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_project.test", "name", updatedProjectName),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "id"),
				),
			},
		},
	})
}

func TestAccProjectResource_import(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Generate project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	require.NoError(t, err)

	// Cleanup project after the test
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_project.test", "name", projectName),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "id"),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "status"),
					resource.TestCheckResourceAttrSet("openaiadmin_project.test", "created_at"),
				),
			},
			{
				ResourceName:      "openaiadmin_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "openaiadmin_project" "test" {
  name = "%s"
}
`, name)
}

func testAccCheckProjectDestroy(client openai.Client) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "openaiadmin_project" {
				continue
			}

			project, err := client.Projects.Retrieve(context.Background(), rs.Primary.ID)
			if err != nil {
				return errors.Wrapf(err, "error retrieving project %s", rs.Primary.ID)
			}
			if project.Status != openai.ProjectStatusArchived {
				return fmt.Errorf("project %s still exists", rs.Primary.ID)
			}
		}
		return nil
	}
}
