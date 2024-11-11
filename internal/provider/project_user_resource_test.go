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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAccProjectUserResource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}

	userID := os.Getenv("OPENAI_TEST_USER_ID")
	resourceName := "openaiadmin_project_user.test"

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectUserDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, string(openai.ProjectUserRoleMember)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectUserExists(ctx, client, resourceName),
					resource.TestCheckResourceAttr(resourceName, "project_id", project.ID),
					resource.TestCheckResourceAttr(resourceName, "user_id", userID),
					resource.TestCheckResourceAttr(resourceName, "role", string(openai.ProjectUserRoleMember)),
					// Verify computed fields are set
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "email"),
					resource.TestCheckResourceAttrSet(resourceName, "added_at"),
				),
			},
			// Import testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s/%s", project.ID, userID),
			},
			// Update testing (role change)
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, string(openai.ProjectUserRoleOwner)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectUserExists(ctx, client, resourceName),
					resource.TestCheckResourceAttr(resourceName, "role", string(openai.ProjectUserRoleOwner)),
				),
			},
		},
	})
}

func TestAccProjectUserResource_InvalidRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}
	userID := os.Getenv("OPENAI_TEST_USER_ID")

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Remove CheckDestroy for invalid role test since the validation will fail during destroy
		Steps: []resource.TestStep{
			// Test creation with invalid role - this should fail at plan time
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, "invalid_role"),
				ExpectError: regexp.MustCompile(
					`Attribute role value must be one of: \["member" "owner"\]`,
				),
			},
		},
	})
}

func TestAccProjectUserResource_UpdateInvalidRole(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}
	userID := os.Getenv("OPENAI_TEST_USER_ID")
	resourceName := "openaiadmin_project_user.test"

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectUserDestroy,
		Steps: []resource.TestStep{
			// First create with valid role
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, string(openai.ProjectUserRoleMember)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectUserExists(ctx, client, resourceName),
				),
			},
			// Try to update with invalid role - this should fail at plan time
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, "invalid_role"),
				ExpectError: regexp.MustCompile(
					`Attribute role value must be one of: \["member" "owner"\]`,
				),
			},
			// Return to valid state for proper cleanup
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, string(openai.ProjectUserRoleMember)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectUserExists(ctx, client, resourceName),
				),
			},
		},
	})
}

func TestAccProjectUserResource_disappears(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Pre-test: Create a project
	projectName := generateTestProject()
	project, err := client.Projects.Create(ctx, projectName)
	if err != nil {
		require.NoError(t, err)
	}
	userID := os.Getenv("OPENAI_TEST_USER_ID")
	resourceName := "openaiadmin_project_user.test"

	// Ensure cleanup after test completion
	t.Cleanup(func() {
		require.NoError(t, client.Projects.Archive(ctx, project.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckProjectUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectUserResourceConfig(project.ID, userID, string(openai.ProjectUserRoleMember)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectUserExists(ctx, client, resourceName),
					// Delete the project user directly using the API client
					func(s *terraform.State) error {
						return client.ProjectUsers.Delete(ctx, project.ID, userID)
					},
				),
				ExpectNonEmptyPlan: true, // Plan should detect the external deletion
			},
		},
	})
}

// Helper function to create the Terraform configuration for testing
func testAccProjectUserResourceConfig(projectID, userID, role string) string {
	return fmt.Sprintf(`
resource "openaiadmin_project_user" "test" {
  project_id = %[1]q
  user_id    = %[2]q
  role       = %[3]q
}
`, projectID, userID, role)
}

// Helper function to verify if the Project User exists in OpenAI
func testAccCheckProjectUserExists(
	ctx context.Context,
	client openai.Client,
	resourceName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Find the resource in state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.Errorf("Project User not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return errors.New("Project User ID is not set")
		}

		// Verify the resource exists in OpenAI
		_, err := client.ProjectUsers.Retrieve(
			ctx,
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["user_id"],
		)
		if err != nil {
			return errors.Wrapf(err, "error retrieving Project User (%s)", rs.Primary.ID)
		}

		return nil
	}
}

// Helper function to verify if the Project User was destroyed
func testAccCheckProjectUserDestroy(s *terraform.State) error {
	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openaiadmin_project_user" {
			continue
		}

		// Try to retrieve the project user
		_, err := client.ProjectUsers.Retrieve(
			ctx,
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["user_id"],
		)
		if err == nil {
			return errors.Errorf("Project User still exists: %s", rs.Primary.ID)
		}

		if !openai.IsNotFoundError(err) {
			return errors.Wrapf(err, "error retrieving Project User (%s)", rs.Primary.ID)
		}
	}

	return nil
}
