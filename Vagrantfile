# -*- mode: ruby -*-
# vi: set ft=ruby :

$ip_file = "db_ip.txt"

Vagrant.configure("2") do |config|
    config.vm.box = 'digital_ocean'
    config.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
    config.ssh.private_key_path = '~/.ssh/keys/digitalocean/digoc_id_rsa'
    config.vm.synced_folder ".", "/vagrant", type: "rsync"
      rsync__exclude = false  # Necessary to include the .env file

    config.vm.define "dbserver", primary: true do |server|
      server.vm.provider :digital_ocean do |provider|
        provider.ssh_key_name = ENV["SSH_KEY_NAME"]
        provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
        provider.image = 'ubuntu-22-04-x64'
        provider.region = 'fra1'
        provider.size = 's-1vcpu-1gb'
        provider.privatenetworking = true
      end
      
      server.vm.hostname = "dbserver"

      server.trigger.after :up do |trigger|
        trigger.info =  "Writing dbserver's IP to file..."
        trigger.ruby do |env,machine|
          remote_ip = machine.instance_variable_get(:@communicator).instance_variable_get(:@connection_ssh_info)[:host]
          File.write($ip_file, remote_ip)
        end
      end

      server.vm.provision "shell", inline: <<-SHELL
        # Fix some issue (as stated in exercises), and run apt-get update
        sudo killall apt apt-get
        sudo rm /var/lib/dpkg/lock-frontend
        sudo apt-get update

        # install http and curl packages
        sudo apt-get install apt-transport-https ca-certificates curl software-properties-common

        # Add Docker's official GPG key:
        sudo apt-get install ca-certificates curl
        sudo install -m 0755 -d /etc/apt/keyrings
        sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
        sudo chmod a+r /etc/apt/keyrings/docker.asc

        # Add the repository to Apt sources:
        echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update

        # Install docker and docker compose
        sudo apt-get install -y docker.io docker-compose-v2

        sudo systemctl status docker
        # sudo usermod -aG docker ${USER}

        # Verify that docker works
        docker run --rm hello-world
        docker rmi hello-world

        # Change directory to the vagrant directory
        cd /vagrant

        # Build postgres docker container
        docker build -t minitwit-postgres -f database/Dockerfile .

        # Run docker container
        docker run -d --name minitwit-postgres-instance -p 5432:5432 -v $(pwd)/database/init:/docker-entrypoint-initdb.d minitwit-postgres

        echo "Database server is running at: $(hostname -I | awk '{print $1}'):5432"
      SHELL
    end

    config.vm.define "webserver", primary: false do |server|
      server.vm.provider :digital_ocean do |provider|
        provider.ssh_key_name = ENV["SSH_KEY_NAME"]
        provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
        provider.image = 'ubuntu-22-04-x64'
        provider.region = 'fra1'
        provider.size = 's-1vcpu-1gb'
        provider.privatenetworking = true
      end

      server.vm.hostname = "webserver"

      server.trigger.before :up do |trigger|
        trigger.info =  "Waiting to create server until dbserver's IP is available."
        trigger.ruby do |env,machine|
          ip_file = "db_ip.txt"
          while !File.file?($ip_file) do
            sleep(1)
          end
          db_ip = File.read($ip_file).strip()
          puts "Now, I have it..."
          puts db_ip
        end
      end

      server.trigger.after :provision do |trigger|
        trigger.ruby do |env,machine|
          File.delete($ip_file) if File.exists? $ip_file
        end
      end

      server.vm.provision "shell", inline: <<-SHELL
        # Fix some issue (as stated in exercises), and run apt-get update
        sudo killall apt apt-get
        sudo rm /var/lib/dpkg/lock-frontend
        sudo apt-get update

        # Get the IP of the database server, and add it to the .env file
        export DB_IP=`cat /vagrant/db_ip.txt`
        echo "\nDB_HOST=$DB_IP" >> /vagrant/.env

        # install http and curl packages
        sudo apt-get install apt-transport-https ca-certificates curl software-properties-common

        # Add Docker's official GPG key:
        sudo apt-get install ca-certificates curl
        sudo install -m 0755 -d /etc/apt/keyrings
        sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
        sudo chmod a+r /etc/apt/keyrings/docker.asc

        # Add the repository to Apt sources:
        echo \
          "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
          $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
          sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update

        # Install docker and docker compose
        sudo apt-get install -y docker.io docker-compose-v2

        sudo systemctl status docker
        # sudo usermod -aG docker ${USER}

        # Verify that docker works
        docker run --rm hello-world
        docker rmi hello-world

        # Change directory to the vagrant directory
        cd /vagrant

        # Build webserver docker container
        docker compose -f docker-compose.yml up -d

        echo "Webserver is running at: http://$(hostname -I | awk '{print $1}'):8080"
        echo "API is running at: http://$(hostname -I | awk '{print $1}'):5000"
      SHELL
    end
end