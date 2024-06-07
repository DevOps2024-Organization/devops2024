# DevOps, Software Evolution and Software Maintenance, Bsc, Spring 2024

## Description

This project is a Twitter-like application designed for the ITU DevOps course, Spring 2024. The application includes essential features such as user registration, login, posting tweets, and viewing timelines. Additionally, it incorporates logging and minitoring using the ELK stack (Elasticsearch, Logstash, and Kibana). The application is containerized using Docker, enabling easy deployment and scalability.

## Report
Please find the project report detailing system perspectives such as architecture, dependencies, and the current state of the systems, as well as process perspectives such as CI/CD processes, monitoring, logging, security assessments, scaling strategies, and lessons learned at this [link](https://github.com/DevOps2024-Organization/devops2024/blob/main/report/build/BSc_group_m.pdf) .

### Key Features of the Project
- User authentication (registration and login)
- Posting and viewing tweets
- Real-time logging and monitoring with the ELK stack
- Dockerized environment for seamless deployment

## Executing program
To run the application, follow these steps:

1. Clone the repository
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

## Running the Application
### Build and start the containers:

```
docker-compose up --build
```
### Access the application:
```
Application: http://localhost:8080
API: http://localhost:5000
Kibana: http://localhost:5601
Grafana: http://localhost:3000
```
## Authors

Daria Damian (dard@itu.dk)

David Zheng (jhou@itu.dk)

Hallgrímur Jónas Jensson (hajj@itu.dk)

Mathias E. L. Rasmussen (memr@itu.dk)

Max-Emil Smith Thorius (maxt@itu.dk)

Fujie Mei (fume@itu.dk)

