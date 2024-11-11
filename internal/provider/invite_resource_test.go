// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/isac322/terraform-provider-openaiadmin/internal/openai"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func generateTestEmail() string {
	return fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
}

func TestAccInviteResource_basic(t *testing.T) {
	t.Parallel()

	email := generateTestEmail()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInviteDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccInviteResourceConfig(email, "reader"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "email", email),
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "role", "reader"),
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "status", "pending"),
					resource.TestCheckResourceAttrSet("openaiadmin_invite.test", "id"),
					resource.TestCheckResourceAttrSet("openaiadmin_invite.test", "invited_at"),
					resource.TestCheckResourceAttrSet("openaiadmin_invite.test", "expires_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "openaiadmin_invite.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update (requires replace) testing
			{
				Config: testAccInviteResourceConfig(email, "owner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "email", email),
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "role", "owner"),
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "status", "pending"),
				),
			},
		},
	})
}

func TestAccInviteResource_externalDeletion(t *testing.T) {
	t.Parallel()

	email := generateTestEmail()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInviteDestroy,
		Steps: []resource.TestStep{
			// First create and verify the resource
			{
				Config: testAccInviteResourceConfig(email, "reader"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_invite.test", "email", email),
					// Delete the resource outside of Terraform
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["openaiadmin_invite.test"]
						if !ok {
							return fmt.Errorf("resource not found in state")
						}
						return client.Invites.Delete(context.Background(), rs.Primary.ID)
					},
				),
			},
			// This step will detect the missing resource during refresh and expect a plan to recreate it
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInviteResource_duplicateEmail(t *testing.T) {
	t.Parallel()

	email := generateTestEmail()
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)

	// Create an invite externally first
	invite, err := client.Invites.Create(context.Background(), email, openai.InviteRoleReader)
	if err != nil {
		require.NoError(t, err)
	}

	// Ensure cleanup of external invite
	t.Cleanup(func() {
		require.NoError(t, client.Invites.Delete(context.Background(), invite.ID))
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckInviteDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInviteResourceConfig(email, "reader"),
				ExpectError: regexp.MustCompile("Error creating invite"), // Should fail due to duplicate email
			},
		},
	})
}

func TestAccInviteResource_multipleRoles(t *testing.T) {
	t.Parallel()

	readerEmail := generateTestEmail()
	ownerEmail := generateTestEmail()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple invites with different roles
			{
				Config: testAccInviteResourceConfigMultiple(readerEmail, ownerEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("openaiadmin_invite.reader", "email", readerEmail),
					resource.TestCheckResourceAttr("openaiadmin_invite.reader", "role", "reader"),
					resource.TestCheckResourceAttr("openaiadmin_invite.owner", "email", ownerEmail),
					resource.TestCheckResourceAttr("openaiadmin_invite.owner", "role", "owner"),
				),
			},
		},
	})
}

func testAccCheckInviteDestroy(s *terraform.State) error {
	client := openai.NewSDKClient(os.Getenv("OPENAI_ADMIN_TOKEN"), nil)
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openaiadmin_invite" {
			continue
		}

		inviteID := rs.Primary.ID
		list, err := client.Invites.List(ctx)
		if err != nil {
			return errors.Wrap(err, "Error listing invites")
		}

		exists := slices.ContainsFunc(list, func(i openai.Invite) bool {
			return i.ID == inviteID
		})
		if exists {
			return errors.Errorf("Invite %s still exists", inviteID)
		}
		return nil
	}

	return nil
}

func testAccInviteResourceConfig(email, role string) string {
	return fmt.Sprintf(`
resource "openaiadmin_invite" "test" {
  email = %[1]q
  role  = %[2]q
}
`, email, role)
}

func testAccInviteResourceConfigMultiple(readerEmail, ownerEmail string) string {
	return fmt.Sprintf(`
resource "openaiadmin_invite" "reader" {
  email = %[1]q
  role  = "reader"
}

resource "openaiadmin_invite" "owner" {
  email = %[2]q
  role  = "owner"
}
`, readerEmail, ownerEmail)
}
