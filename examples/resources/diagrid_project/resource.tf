resource "diagrid_region" "region1" {
  name     = "region1"
  ingress  = "https://*.88288338.xyz:443"
  host     = "regionhost1"
  location = "regionlocation1"
}

resource "diagrid_project" "project" {
  region = diagrid_region.region1.name
  name   = "prj1"
}
