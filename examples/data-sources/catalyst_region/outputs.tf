output "region_name" {
  value = data.catalyst_region.region.name
}

output "region_ingress" {
  value = data.catalyst_region.region.ingress
}

output "region_host" {
  value = data.catalyst_region.region.host
}

output "region_location" {
  value = data.catalyst_region.region.location
}

output "region_join_token" {
  value     = data.catalyst_region.region.join_token
  sensitive = true
}

output "region_connected" {
  value = data.catalyst_region.region.connected
}
