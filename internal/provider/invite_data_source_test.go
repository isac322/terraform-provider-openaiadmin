// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
)

func TestAccInviteDataSource_basic(t *testing.T) {
	t.Parallel()

	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Generate a unique invite for the test
	email := generateTestEmail()

	// Create an invite to test the data source
	invite, err := client.Invites.Create(context.Background(), email, openai.InviteRoleReader)
	if err != nil {
		require.NoError(t, err)
	}

	// Ensure cleanup of the invite after the test
	t.Cleanup(func() {
		require.NoError(t, client.Invites.Delete(context.Background(), invite.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInviteDataSourceConfig(invite.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.openaiadmin_invite.test", "id", invite.ID),
					resource.TestCheckResourceAttr("data.openaiadmin_invite.test", "email", email),
					resource.TestCheckResourceAttr("data.openaiadmin_invite.test", "role", string(openai.InviteRoleReader)),
					resource.TestMatchResourceAttr("data.openaiadmin_invite.test", "status", regexp.MustCompile("(pending|accepted|expired)")),
					resource.TestCheckResourceAttrSet("data.openaiadmin_invite.test", "invited_at"),
					resource.TestCheckResourceAttrSet("data.openaiadmin_invite.test", "expires_at"),
				),
			},
		},
	})
}

func TestAccInviteDataSource_noResults(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccInviteDataSourceConfig("invalid-id"),
				ExpectError: regexp.MustCompile("Invite not found"),
			},
		},
	})
}

func testAccInviteDataSourceConfig(inviteID string) string {
	return fmt.Sprintf(`
data "openaiadmin_invite" "test" {
  id = "%s"
}
`, inviteID)
}
