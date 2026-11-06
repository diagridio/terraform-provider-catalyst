output "organization_name" {
  value = data.diagrid_organization.organization.name
}

output "project_name" {
  value = data.diagrid_project.prj1.name
}

output "project_region" {
  value = data.diagrid_project.prj1.region
}
