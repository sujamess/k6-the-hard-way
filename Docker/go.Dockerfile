# Start from golang base image
FROM golang:1.19.3-alpine3.16 as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Cache the modules
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the neccessary files to build
ARG serviceName
COPY ./pkgs ./pkgs
COPY ./services/${serviceName} ./services/${serviceName}

WORKDIR /app/services/${serviceName}
RUN go build -a -v -o server

FROM scratch

WORKDIR /app
ARG serviceName
COPY --from=builder /app/services/${serviceName}/server ./server

CMD ["./server"]