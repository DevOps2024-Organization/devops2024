# Required providers
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Variables defined in tfvars file
variable "num_workers" {
  default = 1 # Change to the amount of workers that should be instantiated for the replication
}

variable "do_token" {}
variable "ssh_fingerprint" {}

variable "db_port" {}
variable "db_database" {}
variable "db_user" {}
variable "db_pass" {}

# Provider used for the resources
provider "digitalocean" {
  token = var.do_token
}

# Create a droplet for the dbserver
resource "digitalocean_droplet" "dbserver" {
  name          = "dbserver"
  region        = "fra1"
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-1gb"
  ssh_keys      = [var.ssh_fingerprint]

  connection {
    type        = "ssh"
    user        = "root"
    host        = "${self.ipv4_address}"
    private_key = file("~/.ssh/keys/digitalocean/digoc_id_rsa")
  }

  provisioner "file" {
    source      = "${path.module}/database"
    destination = "/root/database"
  }

  provisioner "remote-exec" {
    inline = [
      # DEBIAN_FRONTEND=noninteractive is necessary to avoid promts when apt-get needs to restart services or update configurations 
      "sudo DEBIAN_FRONTEND=noninteractive apt-get update",
      "until sudo DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https ca-certificates curl software-properties-common; do echo 'Dependencies installation failed. Retrying...'; sleep 5; done", # Necessary to include until, since it can include errors if other processes are using it
      "sudo install -m 0755 -d /etc/apt/keyrings",
      "sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
      "sudo chmod a+r /etc/apt/keyrings/docker.asc",
      "echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null",
      "until sudo DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io docker-compose-v2; do echo 'Docker installation failed. Retrying...'; sleep 5; done",
      "if ! sudo systemctl is-active --quiet docker; then",
      "  sudo systemctl start docker",
      "fi",
      "docker run --rm hello-world",
      "docker rmi hello-world",
      "if ! docker volume inspect database-volume &> /dev/null; then",
      "  docker volume create database-volume",
      "fi",
      "docker build -t minitwit-postgres -f database/Dockerfile .",
      "docker run -d --name minitwit-postgres-instance -p ${var.db_port}:5432 -v database-volume:/var/lib/postgresql/data minitwit-postgres"
    ]
  }
}

# Create a droplet for the webserver
resource "digitalocean_droplet" "webserver" {
  depends_on    = [digitalocean_droplet.dbserver]
  name          = "webserver"
  region        = "fra1"
  image         = "ubuntu-22-04-x64"
  size          = "s-4vcpu-8gb"
  ssh_keys      = [var.ssh_fingerprint]

  connection {
    type        = "ssh"
    user        = "root"
    host        = "${self.ipv4_address}"
    private_key = file("~/.ssh/keys/digitalocean/digoc_id_rsa")
  }
  
  provisioner "file" {
    source      = "${path.module}/docker-compose.prod.yml"
    destination = "/root/docker-compose.yml"
  }

  provisioner "file" {
    source      = "${path.module}/prometheus.yml"
    destination = "/root/prometheus.yml"
  }

  provisioner "file" {
    source      = "${path.module}/.htpasswd"
    destination = "/root/.htpasswd"
  }

  provisioner "file" {
    source      = "${path.module}/filebeat.yml"
    destination = "/root/filebeat.yml"
  }

  provisioner "file" {
    source      = "${path.module}/nginx.conf"
    destination = "/root/nginx.conf"
  }

  provisioner "file" {
    source      = "${path.module}/logstash.conf"
    destination = "/root/logstash.conf"
  }

  provisioner "remote-exec" {
    inline = [
      # DEBIAN_FRONTEND=noninteractive is necessary to avoid promts when apt-get needs to restart services or update configurations 
      "sudo DEBIAN_FRONTEND=noninteractive apt-get update",
      "until sudo DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https ca-certificates curl software-properties-common; do echo 'Dependencies installation failed. Retrying...'; sleep 5; done", # Necessary to include until, since it can include errors if other processes are using it
      "sudo install -m 0755 -d /etc/apt/keyrings",
      "sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
      "sudo chmod a+r /etc/apt/keyrings/docker.asc",
      "echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null",
      "until sudo DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io docker-compose-v2; do echo 'Docker installation failed. Retrying...'; sleep 5; done",
      "if ! sudo systemctl is-active --quiet docker; then",
      "  sudo systemctl start docker",
      "fi",
      "docker run --rm hello-world",
      "docker rmi hello-world",
      "echo 'DB_HOST=${digitalocean_droplet.dbserver.ipv4_address}' > .env",
      "echo 'DB_PORT=${var.db_port}' >> .env",
      "echo 'DB_DATABASE=${var.db_database}' >> .env",
      "echo 'DB_USER=${var.db_user}' >> .env",
      "echo 'DB_PASS=${var.db_pass}' >> .env",
      "ufw allow 5000",
      "ufw allow 8080",
      "ufw allow 22/tcp",
      "ufw allow 2376/tcp",
      "ufw allow 2377/tcp",
      "ufw allow 7946/tcp",
      "ufw allow 7946/udp",
      "ufw allow 4789/udp",
      "ufw reload",
      "ufw --force enable",
      "systemctl restart docker",
      "if ! docker info | grep -q \"Swarm: active\"; then",
      "  docker swarm init --advertise-addr ${self.ipv4_address}",
      "fi"
    ]
  }
}

# Create workers for Docker Swarm
resource "digitalocean_droplet" "workers" {
  depends_on    = [digitalocean_droplet.webserver]
  count         = var.num_workers
  name          = "worker${count.index + 1}"
  region        = "fra1"
  image         = "ubuntu-22-04-x64"
  size          = "s-1vcpu-1gb"
  ssh_keys      = [var.ssh_fingerprint]

  connection {
    type        = "ssh"
    user        = "root"
    host        = "${self.ipv4_address}"
    private_key = file("~/.ssh/keys/digitalocean/digoc_id_rsa")
  }

  provisioner "local-exec" {
    command = "ssh -o StrictHostKeyChecking=no -i ~/.ssh/keys/digitalocean/digoc_id_rsa root@${digitalocean_droplet.webserver.ipv4_address} 'docker swarm join-token worker -q' > /tmp/swarm_token.txt"
  }

  provisioner "file" {
    source      = "/tmp/swarm_token.txt"
    destination = "/tmp/swarm_token.txt"
  }

  provisioner "remote-exec" {
    inline = [
      # DEBIAN_FRONTEND=noninteractive is necessary to avoid promts when apt-get needs to restart services or update configurations 
      "sudo DEBIAN_FRONTEND=noninteractive apt-get update",
      "until sudo DEBIAN_FRONTEND=noninteractive apt-get install -y apt-transport-https ca-certificates curl software-properties-common; do echo 'Dependencies installation failed. Retrying...'; sleep 5; done", # Necessary to include until, since it can include errors if other processes are using it
      "sudo install -m 0755 -d /etc/apt/keyrings",
      "sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc",
      "sudo chmod a+r /etc/apt/keyrings/docker.asc",
      "echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null",
      "until sudo DEBIAN_FRONTEND=noninteractiv apt-get install -y docker.io docker-compose-v2; do echo 'Docker installation failed. Retrying...'; sleep 5; done",
      "if ! sudo systemctl is-active --quiet docker; then",
      "  sudo systemctl start docker",
      "fi",
      "docker run --rm hello-world",
      "docker rmi hello-world",
      "if ! docker info | grep -q \"Swarm: active\"; then",
      "  docker swarm join --token $(cat /tmp/swarm_token.txt) ${digitalocean_droplet.webserver.ipv4_address}:2377",
      "fi",
      "rm -f /tmp/swarm-token.txt"
    ]
  }
}

resource "terraform_data" "webserver" {
  depends_on    = [digitalocean_droplet.workers]

  connection {
    type        = "ssh"
    user        = "root"
    host        = "${digitalocean_droplet.webserver.ipv4_address}"
    private_key = file("~/.ssh/keys/digitalocean/digoc_id_rsa")
  }

  provisioner "remote-exec" {
    inline = [
      "docker compose pull",
      "docker stack deploy -c docker-compose.yml minitwit",
      "echo 'Webserver is now running at: http://${digitalocean_droplet.webserver.ipv4_address}:8080",
      "echo 'API is now running at: http://${digitalocean_droplet.webserver.ipv4_address}:5000"
    ]
  }
}