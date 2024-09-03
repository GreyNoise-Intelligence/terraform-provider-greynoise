resource "greynoise_sensor_bootstrap" "this" {
  public_ip =  "44.13.34.10"

  connection {
    host = "44.13.34.10"
    user = "ubuntu"
    port = 22

    private_key = "XXXXX" # private key
  }

  provisioner "remote-exec" {
    inline = [
      self.setup_script,
    ]
  }

  provisioner "remote-exec" {
    inline = [
      # ensure that script can run by waiting for cloud-init to complete
      "cloud-init status --wait > /dev/null",
      self.bootstrap_script,
    ]
    # failure is expected as SSH connection will be lost
    # once bootstrap completes and changes SSH port
    on_failure = continue
  }
}
