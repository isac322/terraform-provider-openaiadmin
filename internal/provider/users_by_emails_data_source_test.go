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

func TestAccUsersByEmailsDataSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Get test user for verification
	userID := os.Getenv("OPENAI_TEST_USER_ID")
	user, err := client.Users.Retrieve(ctx, userID)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with single existing email
			{
				Config: testAccUsersByEmailsDataSourceConfig([]string{user.Email}),
				Check: resource.ComposeTestCheckFunc(
					// Check if the user exists in the map
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.id", user.Email),
						userID,
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.email", user.Email),
						user.Email,
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.role", user.Email),
						string(user.Role),
					),
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.disabled", user.Email),
						fmt.Sprintf("%t", user.Disabled),
					),
					resource.TestCheckResourceAttrSet(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.added_at", user.Email),
					),
				),
			},
			// Test with mix of existing and non-existing emails
			{
				Config: testAccUsersByEmailsDataSourceConfig([]string{
					user.Email,
					"non-existent@example.com",
				}),
				Check: resource.ComposeTestCheckFunc(
					// Check if the existing user is in the map
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						fmt.Sprintf("users.%s.id", user.Email),
						userID,
					),
					// Verify the map size (should only contain the existing user)
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						"users.%",
						"1",
					),
				),
			},
			// Test with only non-existing emails
			{
				Config: testAccUsersByEmailsDataSourceConfig([]string{
					"non-existent1@example.com",
					"non-existent2@example.com",
				}),
				Check: resource.ComposeTestCheckFunc(
					// Verify the map is empty
					resource.TestCheckResourceAttr(
						"data.openaiadmin_users_by_emails.test",
						"users.%",
						"0",
					),
				),
			},
		},
	})
}

func testAccUsersByEmailsDataSourceConfig(emails []string) string {
	// Convert emails array to HCL list string
	emailsList := "["
	for i, email := range emails {
		if i > 0 {
			emailsList += ", "
		}
		emailsList += fmt.Sprintf("%q", email)
	}
	emailsList += "]"

	return fmt.Sprintf(`
data "openaiadmin_users_by_emails" "test" {
  emails = %s
}
`, emailsList)
}
