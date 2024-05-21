FROM docker.elastic.co/logstash/logstash-oss:7.1.0

# Create a non-root user and group
RUN addgroup -S logstash && adduser -S -G logstash logstash

# Change ownership of the logstash configuration directory
RUN chown -R logstash:logstash /usr/share/logstash

# Switch to the non-root user
USER logstash

COPY ./logstash.conf /etc/logstash/conf.d/

CMD ["logstash", "-f", "/etc/logstash/conf.d/logstash.conf"]