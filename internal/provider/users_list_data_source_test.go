// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/stretchr/testify/require"
)

func TestAccUsersListDataSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Get test user for verification
	userID := os.Getenv("OPENAI_TEST_USER_ID")
	user, err := client.Users.Retrieve(ctx, userID)
	require.NoError(t, err)

	// Get full list for count verification
	allUsers, err := client.Users.List(ctx)
	require.NoError(t, err)
	userCount := len(allUsers)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUsersListDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					// Verify list length
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_list.test",
						"users.#",
						fmt.Sprintf("%d", userCount),
					),
					// Verify each user has required attributes
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_users_list.test",
						"users.0.id",
					),
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_users_list.test",
						"users.0.email",
					),
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_users_list.test",
						"users.0.role",
					),
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_users_list.test",
						"users.0.added_at",
					),
					// Verify known test user exists in the list
					testAccCheckUsersListContainsUser(userID, user.Email, string(user.Role), user.Disabled),
				),
			},
		},
	})
}

func testAccCheckUsersListContainsUser(id, email, role string, disabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.openaiadmin_users_list.test"]
		if !ok {
			return fmt.Errorf("resource not found in state")
		}

		userCount, err := strconv.Atoi(rs.Primary.Attributes["users.#"])
		if err != nil {
			return err
		}

		// Look for the user in the list
		for i := 0; i < userCount; i++ {
			prefix := fmt.Sprintf("users.%d.", i)
			if rs.Primary.Attributes[prefix+"id"] == id {
				if rs.Primary.Attributes[prefix+"email"] != email {
					return fmt.Errorf("email mismatch for user %s", id)
				}
				if rs.Primary.Attributes[prefix+"role"] != role {
					return fmt.Errorf("role mismatch for user %s", id)
				}
				if rs.Primary.Attributes[prefix+"disabled"] != fmt.Sprintf("%t", disabled) {
					return fmt.Errorf("disabled status mismatch for user %s", id)
				}
				return nil
			}
		}

		return fmt.Errorf("user %s not found in the list", id)
	}
}

func testAccUsersListDataSourceConfig() string {
	return `
data "openaiadmin_users_list" "test" {
}
`
}
