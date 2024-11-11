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
)

func TestAccUserResource(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	userID := os.Getenv("OPENAI_TEST_USER_ID")
	resourceName := "openaiadmin_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories:  testAccProtoV6ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			// Verify create operation is not supported
			{
				Config: testAccUserResourceConfig_empty(),
				ExpectError: regexp.MustCompile(
					"User creation is not supported by this resource",
				),
			},
			// Update role testing
			{
				Config:        testAccUserResourceConfig_withRole(string(openai.UserRoleOwner)),
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: userID,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, client, resourceName),
					resource.TestCheckResourceAttr(resourceName, "role", string(openai.UserRoleOwner)),
				),
			},
		},
	})
}

func testAccUserResourceConfig_empty() string {
	return `
resource "openaiadmin_user" "test" {
}`
}

func testAccUserResourceConfig_withRole(role string) string {
	return fmt.Sprintf(`
resource "openaiadmin_user" "test" {
  role = %[1]q
}
`, role)
}

func testAccCheckUserExists(ctx context.Context, client openai.Client, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return errors.Errorf("User not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return errors.New("User ID is not set")
		}

		_, err := client.Users.Retrieve(ctx, rs.Primary.ID)
		if err != nil {
			return errors.Wrapf(err, "error retrieving User (%s)", rs.Primary.ID)
		}

		return nil
	}
}
