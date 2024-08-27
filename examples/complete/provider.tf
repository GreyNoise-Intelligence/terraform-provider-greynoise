terraform {
  required_providers {
    greynoise = {
      source  = "greynoise-io/greynoise"
      version = "1.0.0"
    }
  }
}

provider "greynoise" {
  // GN_API_KEY env var is used to provide key
}

data "greynoise_personas" "rdp" {
  search = "rdp"
  limit  = 1
}

resource "greynoise_sensor_bootstrap" "this" {
  public_ip = var.sensor_ip

  connection {
    host = var.sensor_ip
    user = var.sensor_ssh.user
    port = var.sensor_ssh.port

    private_key = file(var.sensor_ssh.ssh_key_file)
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
