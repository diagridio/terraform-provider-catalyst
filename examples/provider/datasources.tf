data "catalyst_organization" "organization" {
  id = var.organization_id
}

data "catalyst_project" "prj1" {
  organization_id = data.catalyst_organization.organization.id
  name            = var.project_name
}

