---
page_title: "simplemdm_scriptjob Resource - terraform-provider-simplemdm"
subcategory: ""
description: |-
  Manages script job executions in SimpleMDM for running scripts on targeted macOS devices.
---

# simplemdm_scriptjob (Resource)

Manages script job executions in SimpleMDM. Script jobs execute scripts on targeted macOS devices, allowing you to automate device management tasks, gather information, or perform maintenance operations at scale.

## Example Usage

### Basic Example

```terraform
resource "simplemdm_script" "hello" {
  name       = "Hello Script"
  scriptfile = "#!/bin/bash\necho 'Hello World'"
}

resource "simplemdm_scriptjob" "run_hello" {
  script_id  = simplemdm_script.hello.id
  device_ids = ["123456"]
}
```

### Advanced Example - Target Multiple Devices

```terraform
resource "simplemdm_script" "update_software" {
  name       = "Update Software"
  scriptfile = file("${path.module}/scripts/update.sh")
}

resource "simplemdm_scriptjob" "update_all" {
  script_id  = simplemdm_script.update_software.id
  device_ids = ["123456", "234567", "345678"]
}
```

### Advanced Example - Target Device Groups

```terraform
resource "simplemdm_script" "collect_logs" {
  name       = "Collect System Logs"
  scriptfile = file("${path.module}/scripts/collect_logs.sh")
}

resource "simplemdm_scriptjob" "logs_from_engineering" {
  script_id = simplemdm_script.collect_logs.id
  group_ids = ["123456", "234567"]  # Device group IDs
}
```

### Advanced Example - Target Assignment Groups

```terraform
resource "simplemdm_script" "security_audit" {
  name       = "Security Audit"
  scriptfile = file("${path.module}/scripts/audit.sh")
}

resource "simplemdm_scriptjob" "audit_executives" {
  script_id            = simplemdm_script.security_audit.id
  assignment_group_ids = ["789012"]
}
```

### Advanced Example - Store Output in Custom Attribute

```terraform
resource "simplemdm_attribute" "last_backup_status" {
  name = "last_backup_status"
}

resource "simplemdm_script" "backup" {
  name       = "Backup Script"
  scriptfile = file("${path.module}/scripts/backup.sh")
}

resource "simplemdm_scriptjob" "backup_job" {
  script_id              = simplemdm_script.backup.id
  device_ids             = ["123456"]
  custom_attribute       = "last_backup_status"
  custom_attribute_regex = "\\n"  # Clean up newlines from output
}
```

### Advanced Example - Multiple Target Types

```terraform
resource "simplemdm_script" "inventory" {
  name       = "Device Inventory"
  scriptfile = file("${path.module}/scripts/inventory.sh")
}

resource "simplemdm_scriptjob" "comprehensive_inventory" {
  script_id = simplemdm_script.inventory.id
  
  # Target specific devices
  device_ids = ["111111", "222222"]
  
  # Target device groups
  group_ids = ["333333"]
  
  # Target assignment groups
  assignment_group_ids = ["444444"]
  
  # Store results
  custom_attribute       = "inventory_data"
  custom_attribute_regex = "\\n"
}
```

### Advanced Example - Monitoring Job Status

```terraform
resource "simplemdm_script" "health_check" {
  name       = "System Health Check"
  scriptfile = file("${path.module}/scripts/health_check.sh")
}

resource "simplemdm_scriptjob" "health_monitoring" {
  script_id  = simplemdm_script.health_check.id
  device_ids = ["123456", "234567"]
}

# Output job status for monitoring
output "job_status" {
  value = {
    job_id         = simplemdm_scriptjob.health_monitoring.id
    status         = simplemdm_scriptjob.health_monitoring.status
    success_count  = simplemdm_scriptjob.health_monitoring.success_count
    errored_count  = simplemdm_scriptjob.health_monitoring.errored_count
    pending_count  = simplemdm_scriptjob.health_monitoring.pending_count
  }
}
```

### Advanced Example - Dynamic Device Targeting

```terraform
data "simplemdm_devices" "macos_devices" {
  search = "Mac"
}

resource "simplemdm_script" "macos_config" {
  name       = "macOS Configuration"
  scriptfile = file("${path.module}/scripts/macos_config.sh")
}

resource "simplemdm_scriptjob" "configure_all_macs" {
  script_id = simplemdm_script.macos_config.id
  
  device_ids = [
    for device in data.simplemdm_devices.macos_devices.devices : 
    device.id
  ]
}
```

### Advanced Example - Conditional Execution

```terraform
variable "run_maintenance" {
  type    = bool
  default = false
}

resource "simplemdm_script" "maintenance" {
  name       = "System Maintenance"
  scriptfile = file("${path.module}/scripts/maintenance.sh")
}

resource "simplemdm_scriptjob" "conditional_maintenance" {
  count = var.run_maintenance ? 1 : 0
  
  script_id  = simplemdm_script.maintenance.id
  device_ids = ["123456"]
}
```

### Advanced Example - Accessing Device Results

```terraform
resource "simplemdm_script" "disk_space" {
  name       = "Check Disk Space"
  scriptfile = "#!/bin/bash\ndf -h | grep '/System/Volumes/Data'"
}

resource "simplemdm_scriptjob" "check_space" {
  script_id  = simplemdm_script.disk_space.id
  device_ids = ["123456", "234567"]
}

# Output individual device results
output "device_responses" {
  value = {
    for device in simplemdm_scriptjob.check_space.devices :
    device.id => {
      status   = device.status
      response = device.response
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `script_id` (String) The ID of the script to be run on the devices.

### Optional

- `assignment_group_ids` (Set of String) A list of assignment group IDs to run the script on. All macOS devices from these assignment groups will be included. At least one of `device_ids`, `group_ids`, or `assignment_group_ids` must be provided.
- `custom_attribute` (String) If provided, the output from the script will be stored in this custom attribute on each device.
- `custom_attribute_regex` (String) Used to sanitize the output from the script before storing it in the custom attribute. Can be left empty but `\\n` is recommended to remove newlines.
- `device_ids` (Set of String) A list of device IDs to run the script on. At least one of `device_ids`, `group_ids`, or `assignment_group_ids` must be provided.
- `group_ids` (Set of String) A list of device group IDs to run the script on. All macOS devices from these groups will be included. At least one of `device_ids`, `group_ids`, or `assignment_group_ids` must be provided.

### Read-Only

- `content` (String) Script contents that were executed by the job.
- `created_at` (String) Creation timestamp returned by the API.
- `created_by` (String) User or API key that created the job.
- `devices` (Attributes List) Execution results for each targeted device. (see [below for nested schema](#nestedatt--devices))
- `errored_count` (Number) Number of devices that failed to execute the script.
- `id` (String) The unique identifier of the script job in SimpleMDM.
- `job_identifier` (String) Identifier reported by the SimpleMDM API for the job.
- `job_name` (String) Human friendly name of the job.
- `pending_count` (Number) Number of devices that have not yet reported a result.
- `script_name` (String) Name of the script that was executed.
- `status` (String) Current execution status of the job.
- `success_count` (Number) Number of devices that completed successfully.
- `updated_at` (String) Last update timestamp returned by the API.
- `variable_support` (Boolean) Indicates whether the script supports variables.

<a id="nestedatt--devices"></a>
### Nested Schema for `devices`

Read-Only:

- `id` (String) Device identifier.
- `response` (String) Output returned by the device, when available.
- `status` (String) Execution status reported for the device.
- `status_code` (String) Optional status code returned by the device.

## Import

Import is supported using the following syntax:

```shell
# Script job can be imported by specifying the job ID
terraform import simplemdm_scriptjob.example 123456
```

## Notes

- **Async Execution**: Script jobs execute asynchronously. The resource is created when the job is queued, not when execution completes.
- **Target Requirements**: You must specify at least one of `device_ids`, `group_ids`, or `assignment_group_ids`.
- **macOS Only**: Scripts are only executed on macOS devices. Non-macOS devices in specified groups are automatically skipped.
- **Output Storage**: Use `custom_attribute` to store script output on each device. This is useful for inventory collection or status reporting.
- **Regex Sanitization**: The `custom_attribute_regex` parameter helps clean up script output before storage. Common patterns:
  - `\\n` - Remove newlines
  - `\\s+` - Remove extra whitespace
  - Custom patterns for extracting specific data
- **Job Status**: Monitor execution via the `status`, `success_count`, `errored_count`, and `pending_count` attributes.
- **Device Results**: Access individual device execution results via the `devices` attribute for detailed troubleshooting.
- **Idempotency**: Creating the same script job multiple times will execute it multiple times. Use lifecycle rules if needed.
- **Variable Support**: If the script has `variablesupport = true`, custom attributes will be substituted before execution.
- **Error Handling**: Failed executions are tracked in `errored_count` and individual device status can be inspected.
- **Best Practices**:
  - Test scripts on a small device set first
  - Use custom attributes to track execution results
  - Monitor job status and device responses
  - Handle script errors gracefully
  - Consider device connectivity when scheduling