# Bind an existing app to an existing assignment group that is managed in the
# SimpleMDM UI — typically a dynamic assignment group whose membership rules
# are not exposed via the API.
resource "simplemdm_assignmentgroup_app_binding" "example" {
  assignment_group_id = var.assignment_group_id
  app_id              = simplemdm_app.example.id

  # Optional overrides — only meaningful for Munki-managed apps.
  deployment_type = "standard"
}

resource "simplemdm_app" "example" {
  app_store_id = "284882215"
}

variable "assignment_group_id" {
  description = "Identifier of the assignment group to bind the app to."
  type        = string
}
