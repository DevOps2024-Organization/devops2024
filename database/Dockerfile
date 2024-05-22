FROM postgres:latest

ARG db
ARG dbuser
ARG dbpassword

# Set environmental variables
ENV POSTGRES_DB=$db
ENV POSTGRES_USER=$dbuser
ENV POSTGRES_PASSWORD=$dbpassword


# Create a new non-root user and set password
RUN useradd -m -d /var/lib/postgresql minitwit_user && \
    echo "minitwit_user:secure_password" | chpasswd && \
    chown -R minitwit_user:postgres /var/lib/postgresql && \
    chown -R minitwit_user:postgres /var/run/postgresql && \
    chown -R minitwit_user:postgres /usr/lib/postgresql
# Switch to new user
USER minitwit_user

# Expose Postgres port
EXPOSE 5432

CMD ["postgres"]
