variable "sensor_ip" {
  description = "Public IP of the sensor"
  type        = string
}

variable "sensor_ssh" {
  description = "SSH parameters"
  type        = object({
    user    = string
    port    = number
    ssh_key_file = string
  })
}
