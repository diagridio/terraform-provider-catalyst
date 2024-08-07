data "catalyst_organization" "current" {}

data "catalyst_region" "onebox" {
  name = "Onebox Host #1"
}

data "catalyst_project" "project" {
  organization_id = data.catalyst_organization.current.id
  region          = data.catalyst_region.onebox.id
  name            = "prj1"
}

