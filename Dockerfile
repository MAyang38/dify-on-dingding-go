# Use the official Golang image as a base image
FROM golang:1.21-alpine


ENV GOPROXY=https://goproxy.cn,direct

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 7777

# Command to run the executable
CMD ["./main"]
