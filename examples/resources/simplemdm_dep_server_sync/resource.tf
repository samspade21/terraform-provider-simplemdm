resource "simplemdm_dep_server_sync" "manual" {
  dep_server_id = "1"
  triggers = {
    # Change any value to retrigger the sync.
    nonce = "2025-05-06"
  }
}
