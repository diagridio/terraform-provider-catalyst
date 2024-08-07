resource "catalyst_project" "project" {
  organization_id = data.catalyst_organization.current.id
  region          = data.catalyst_region.current.id
  name            = "prj1"

  managed_workflow = true
  managed_pubsub   = false
  managed_kvstore  = false
}
