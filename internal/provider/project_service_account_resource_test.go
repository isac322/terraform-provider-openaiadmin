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

func TestAccProjectServiceAccountResource(t *testing.T) {
	t.Parallel()

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	ctx := context.Background()
	projectName := generateTestProject()

	// Pre-test: Create an API Key
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})
	resourceName := "openaiadmin_project_service_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectServiceAccountDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectServiceAccountResourceConfig(project.ID, "test-acc-sa"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-sa"),
					resource.TestCheckResourceAttr(resourceName, "role", "member"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					// API Key checks
					resource.TestCheckResourceAttrSet(resourceName, "api_key.value"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.id"),
				),
			},
			// Update testing (requires replace)
			{
				Config: testAccProjectServiceAccountResourceConfig(project.ID, "test-acc-sa-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-acc-sa-updated"),
					resource.TestCheckResourceAttr(resourceName, "role", "member"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					// API Key checks for new instance
					resource.TestCheckResourceAttrSet(resourceName, "api_key.value"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccProjectServiceAccountResource_disappears(t *testing.T) {
	t.Parallel()

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	ctx := context.Background()
	projectName := generateTestProject()

	// Pre-test: Create an API Key
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})
	resourceName := "openaiadmin_project_service_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectServiceAccountResourceConfig(project.ID, "test-acc-sa-disappears"),
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return errors.Errorf("Project Service Account not found: %s", resourceName)
					}

					if rs.Primary.ID == "" {
						return errors.New("Project Service Account ID is not set")
					}

					return client.ProjectServiceAccounts.Delete(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID)
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccProjectServiceAccountResourceConfig(projectId, name string) string {
	return fmt.Sprintf(`
resource "openaiadmin_project_service_account" "test" {
  project_id = %[1]q
  name       = %[2]q
}
`, projectId, name)
}

func testAccCheckProjectServiceAccountDestroy(s *terraform.State) error {
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openaiadmin_project_service_account" {
			continue
		}

		if rs.Primary.ID == "" {
			return errors.New("No Project Service Account ID is set")
		}

		_, err := client.ProjectServiceAccounts.Retrieve(ctx, rs.Primary.Attributes["project_id"], rs.Primary.ID)
		if err == nil {
			return errors.New("Project Service Account still exists")
		}
		if openai.IsNotFoundError(err) {
			return nil
		}

		return errors.Wrapf(err, "Error retrieving Project Service Account %s", rs.Primary.ID)
	}

	return nil
}
