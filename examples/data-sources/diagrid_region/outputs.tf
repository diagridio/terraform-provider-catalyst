output "region_name" {
  value = data.diagrid_region.region.name
}

output "region_ingress" {
  value = data.diagrid_region.region.ingress
}

output "region_host" {
  value = data.diagrid_region.region.host
}

output "region_location" {
  value = data.diagrid_region.region.location
}

output "region_join_token" {
  value     = data.diagrid_region.region.join_token
  sensitive = true
}

output "region_connected" {
  value = data.diagrid_region.region.connected
}
