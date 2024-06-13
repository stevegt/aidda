# Use an official Go image as the base image
FROM golang:1.22 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the ./remote/ directory to the Working Directory inside the container
COPY ./remote/ .

# Build the Go app
RUN go build -o actionRunner actionRunner.go

# Command to run the executable
CMD ["./actionRunner"]

