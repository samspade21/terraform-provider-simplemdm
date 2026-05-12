# List all assignment groups and filter
data "simplemdm_assignmentgroups" "all" {}

# Filter by auto-deploy enabled
output "auto_deploy_groups" {
  value = [for group in data.simplemdm_assignmentgroups.all.assignment_groups : group.name if group.auto_deploy]
}

# Filter by app-track-location enabled
output "location_tracking_groups" {
  value = [for group in data.simplemdm_assignmentgroups.all.assignment_groups : group.name if group.app_track_location]
}

# Get groups sorted by priority
output "groups_by_priority" {
  value = [for group in sort([for g in data.simplemdm_assignmentgroups.all.assignment_groups : {
    name     = g.name
    priority = g.priority
  }], "priority") : group.name]
}