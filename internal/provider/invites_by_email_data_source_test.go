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

func TestAccInvitesByEmailDataSource_basic(t *testing.T) {
	t.Parallel()

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Generate a unique email for testing
	email := generateTestEmail()

	// Create a pending invite
	invitePending, err := client.Invites.Create(context.Background(), email, openai.InviteRoleReader)
	if err != nil {
		require.NoError(t, err)
	}

	// Ensure cleanup of the invite
	t.Cleanup(func() {
		require.NoError(t, client.Invites.Delete(context.Background(), invitePending.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvitesByEmailDataSourceConfig(email, "pending"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_invites_by_email.test", "email", email),
					resource.TestCheckResourceAttrSet("data.openaiadmin_invites_by_email.test", "invites.0.id"),
					resource.TestCheckResourceAttr("data.openaiadmin_invites_by_email.test", "invites.0.status", "pending"),
				),
			},
		},
	})
}

func TestAccInvitesByEmailDataSource_noResults(t *testing.T) {
	t.Parallel()

	email := generateTestEmail()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvitesByEmailDataSourceConfig(email, "pending"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_invites_by_email.test", "email", email),
					resource.TestCheckResourceAttr("data.openaiadmin_invites_by_email.test", "invites.#", "0"), // Ensure invites is an empty list
				),
			},
		},
	})
}

func testAccInvitesByEmailDataSourceConfig(email, status string) string {
	return fmt.Sprintf(`
data "openaiadmin_invites_by_email" "test" {
  email  = "%s"
  status = "%s"
}
`, email, status)
}
