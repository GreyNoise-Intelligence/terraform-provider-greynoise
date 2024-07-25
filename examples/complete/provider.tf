terraform {
  required_providers {
    greynoise = {
      source  = "greynoise-io/greynoise"
      version = "1.0.0"
    }
  }
}

provider "greynoise" {
  workspace_id = "75a76a71-5cc1-492c-a8b7-b546bd4959ae"
}

data "greynoise_personas" "rdp" {
  search = "rdp"
  limit  = 1
}

resource "greynoise_sensor_bootstrap" "this" {
  public_ip = "44.202.75.6"

  connection {
    host        = "44.202.75.6"
    port        = 22
    user        = "ubuntu"
    private_key = file("~/.terraform.d/nayyara-sensors.pem")
  }

  provisioner "remote-exec" {
    inline = [
      self.setup_script,
    ]
  }

  provisioner "remote-exec" {
    inline = [
      self.bootstrap_script,
    ]
    # failure is expected as SSH connection will be lost
    # once bootstrap completes and changes SSH port
    on_failure = continue
  }
}

resource "greynoise_sensor" "this" {
  boostrap_connection = {
    host        = "44.202.75.6"
    port        = greynoise_sensor_bootstrap.this.ssh_port_selected
    user        = "ubuntu"
    private_key = file("~/.terraform.d/nayyara-sensors.pem")
  }

}

output "personas" {
  value = {
    ids = data.greynoise_personas.rdp.ids
  }
}

output "sensor_bootstrap" {
  value = greynoise_sensor_bootstrap.this.ssh_port_selected
}
