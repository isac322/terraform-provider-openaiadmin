---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openaiadmin_project_user Resource - openaiadmin"
subcategory: ""
description: |-
  Project User resource
---

# openaiadmin_project_user (Resource)

Project User resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) The ID of the project to which this user belongs.
- `role` (String) The role of the project user.
- `user_id` (String) The ID of the user to be added to the project.

### Read-Only

- `added_at` (String) The timestamp when the user was added to the project.
- `email` (String) The email of the project user.
- `id` (String) The ID of the project user. Format: `{project_id}/{user_id}`
- `name` (String) The name of the project user.
