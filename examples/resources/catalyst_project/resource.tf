resource "catalyst_region" "region1" {
  name     = "region1"
  ingress  = "https://*.88288338.xyz:443"
  host     = "regionhost1"
  location = "regionlocation1"
}

resource "catalyst_project" "project" {
  # region          = data.catalyst_region.onebox.name
  region = catalyst_region.region1.name
  name   = "prj1"
}
