# Use the same build stage as before to compile your application
FROM golang:1.18 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o minitwit ./src/main.go
RUN chmod +x minitwit

# Start of the final stage
FROM alpine:latest as final

# Install ca-certificates and netcat
RUN apk --no-cache add ca-certificates netcat-openbsd tzdata && \
  cp /usr/share/zoneinfo/UTC /etc/localtime && \
  echo "UTC" > /etc/timezone && \
  apk del tzdata

# Set environment variables for locale, if needed
ENV LANG en_US.UTF-8  
ENV LANGUAGE en_US:en  
ENV LC_ALL en_US.UTF-8  

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/minitwit .

# Copy other necessary files
COPY --from=builder /app/.env .
COPY --from=builder /app/src/web/templates ./src/web/templates
COPY --from=builder /app/src/web/static ./src/web/static

# Expose the port the app runs on
EXPOSE 8080

CMD ["./minitwit"]