# set an output for the org id
output "organization_id" {
  value = data.catalyst_organization.current.id
}
