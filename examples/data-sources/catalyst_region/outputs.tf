# set an output for the region id
output "region_id" {
  value = data.catalyst_region.current.id
}
