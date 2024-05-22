FROM postgres:latest

ARG db
ARG dbuser
ARG dbpassword

# Set environmental variables
ENV POSTGRES_DB=$db
ENV POSTGRES_USER=$dbuser
ENV POSTGRES_PASSWORD=$dbpassword

# Expose Postgres port
EXPOSE 5432

CMD ["postgres"]
