output "organization_name" {
  value = data.catalyst_organization.organization.name
}

output "project_name" {
  value = data.catalyst_project.prj1.name
}

output "project_region" {
  value = data.catalyst_project.prj1.region
}
