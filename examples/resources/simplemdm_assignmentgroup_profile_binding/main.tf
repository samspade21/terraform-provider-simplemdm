# Bind an existing custom configuration profile (or custom declaration) to an
# existing assignment group that is managed in the SimpleMDM UI — typically a
# dynamic assignment group whose membership rules are not exposed via the API.
resource "simplemdm_assignmentgroup_profile_binding" "example" {
  assignment_group_id = var.assignment_group_id
  profile_id          = simplemdm_customprofile.example.id
}

resource "simplemdm_customprofile" "example" {
  name         = "Example Profile"
  mobileconfig = file("${path.module}/example.mobileconfig")
}

variable "assignment_group_id" {
  description = "Identifier of the assignment group to bind the profile to."
  type        = string
}
