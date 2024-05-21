# Build Go API
FROM golang:latest as builder

# Set Current Working Directory inside the container
WORKDIR /app

# Copy mod and sum files
COPY go.mod go.sum ./

# Download Dependencies
RUN go mod download

# Copy source from current to Working Directory
COPY . .
COPY .env /app/


# Build Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./api/api.go

# Setup runtime environment
FROM alpine:latest

# Create a non-root user
RUN adduser -D nonroot


# Set current working Directoy
WORKDIR /root/nonroot/

# Copy pre-built binary file
COPY --from=builder /app/api .
COPY --from=builder /app/.env .


# Change change ownership of files to non-root user
RUN chown -R nonroot:nonroot /home/nonroot

# Switch to non-root user
USER nonroot

# Expose port 8080
EXPOSE 5000

# Run the excutable
CMD ["./api"]
