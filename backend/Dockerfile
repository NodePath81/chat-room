FROM golang:1.23.6

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y gcc libc6-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"] 