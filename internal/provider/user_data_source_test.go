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

func TestAccUserDataSource(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	userID := os.Getenv("OPENAI_TEST_USER_ID")

	// Get user info for verification
	user, err := client.Users.Retrieve(ctx, userID)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig(userID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_user.test", "id", userID),
					resource.TestCheckResourceAttr("data.openaiadmin_user.test", "email", user.Email),
					resource.TestCheckResourceAttr("data.openaiadmin_user.test", "role", string(user.Role)),
					resource.TestCheckResourceAttr("data.openaiadmin_user.test", "disabled", fmt.Sprintf("%t", user.Disabled)),
					resource.TestCheckResourceAttrSet("data.openaiadmin_user.test", "added_at"),
				),
			},
		},
	})
}

func TestAccUserDataSource_NonExistent(t *testing.T) {
	if os.Getenv("ENV") == "local" {
		t.Parallel()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig("non-existent-user-id"),
				ExpectError: regexp.MustCompile(
					"Error reading user",
				),
			},
		},
	})
}

func testAccUserDataSourceConfig(userID string) string {
	return fmt.Sprintf(`
data "openaiadmin_user" "test" {
  id = %[1]q
}
`, userID)
}
