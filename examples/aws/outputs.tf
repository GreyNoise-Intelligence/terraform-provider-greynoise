output "personas" {
  description = "RDP personas"
  value       = {
    ids = data.greynoise_personas.rdp.ids
  }
}

output "sensor" {
  description = "Sensor information"
  value       = {
    public_ip = aws_instance.this.public_ip
    ssh_port  = greynoise_sensor_bootstrap.this.ssh_port_selected
  }
}
