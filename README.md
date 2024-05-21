# ITU DevOps course, Spring 2024

## Description

This project is a Twitter-like application designed for the ITU DevOps course, Spring 2024. The application includes essential features such as user registration, login, posting tweets, and viewing timelines. Additionally, it incorporates logging and minitoring using the ELK stack (Elasticsearch, Logstash, and Kibana). The application is containerized using Docker, enabling easy deployment and scalability.

### Key Features
- User authentication (registration and login)
- Posting and viewing tweets
- Real-time logging and monitoring with the ELK stack
- Dockerized environment for seamless deployment

## Executing program
To run the application, follow these steps:

1. ** Clone the repository**
```
git clone https://github.com/DevOps2024-Organization/devops2024.git
```
2. Set up Enviroment Variables

In root directory of source code in the terraform.tfvars file set the following variables:

```
do_token           = "{Digital Ocean API Token}"

ssh_fingerprint    = "{Fingerprint of Registered SSH Key on Digital Ocean}"

num_workers        = {1 || 2}

db_port            = "{database port}"

db_database        = "{database name}"

db_user            = "{database user}"

db_pass            = "{database password}"

docker_username    = "{Docker Username}"

docker_password    = "{Docker Password / Access Token}"
```

3. run
```
terraform plan
```

```
terraform apply
```

## Authors

dard@itu.dk

hajj@itu.dk

memr@itu.dk

maxt@itu.dk

fume@itu.dk



