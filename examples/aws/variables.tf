variable "vpc" {
  description = "VPC parameters"
  type        = object({
    vpc_id    = string
    subnet_id = string
  })
}

variable "key_pair" {
  description = "Key pair for EC2 instance SSH"
  type        = object({
    name        = string
    private_key_file = string
  })
}
