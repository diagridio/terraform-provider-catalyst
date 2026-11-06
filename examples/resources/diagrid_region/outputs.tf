output "region_name" {
  value = diagrid_region.region.name
}

output "region_ingress" {
  value = diagrid_region.region.ingress
}

output "region_host" {
  value = diagrid_region.region.host
}

output "region_location" {
  value = diagrid_region.region.location
}

output "region_type" {
  value = diagrid_region.region.type
}

output "region_join_token" {
  value     = diagrid_region.region.join_token
  sensitive = true
}

output "region_connected" {
  value = diagrid_region.region.connected
}

