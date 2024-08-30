output "personas" {
  description = "RDP personas"
  value       = {
    ids = data.greynoise_personas.rdp.ids
  }
}

output "sensor_bootstrap" {
  description = "Sensor bootstrap parameters"
  value       = {
    ssh_port = greynoise_sensor_bootstrap.this.ssh_port_selected
  }
}
