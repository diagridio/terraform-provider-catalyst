output "region_name" {
  value = catalyst_region.region.name
}

output "region_ingress" {
  value = catalyst_region.region.ingress
}

output "region_host" {
  value = catalyst_region.region.host
}

output "region_location" {
  value = catalyst_region.region.location
}

output "region_type" {
  value = catalyst_region.region.type
}

output "region_join_token" {
  value = catalyst_region.region.join_token
  sensitive = true
}

output "region_connected" {
  value = catalyst_region.region.connected
}

