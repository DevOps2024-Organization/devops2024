# ITU DevOps course, Spring 2024

## Run On Local Machine
Make sure you have a `.env` file in the root folder, and that it contains the following fields:
```
# PostgreSQL database configuration
DB_HOST=localhost
DB_PORT=5433
DB_DATABASE=minitwit
DB_USER=root
DB_PASS=password
```
Then run the following commands:
```
make postgresinit
make createdb
make run
```

## Docker
Make sure you have a `.env` file in the root folder, and that it contains the following fields:
```
# PostgreSQL database configuration
DB_HOST=db
DB_PORT=5432
DB_DATABASE=minitwit
DB_USER=root
DB_PASS=password
```
Start service:
```
docker-compose build
docker-compose up -d
```

Stop service:
```
docker-compose down
```

Stop service and remove container:
```
docker-compose down --volumes
```

## API 23/2/2024
Created API version
Passed 9/9 tests

## Digital Ocean Deployment 27/2/2024
Deployed to Digital Ocean
web: http://104.248.43.157:8080/
API: http://104.248.43.157:5000/

- list of commands: ./docker_notes
- By running those commands in order, first on the dbserver then on the webserver they will run containers and the webserver will connect to the database.

// builds and runs db, app and api locally with docker
local_minitwit_build_and_run.sh

// stops and removes containers and images so you can build clean:
docker_cleanup.sh