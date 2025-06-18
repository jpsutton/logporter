# Build logporter
FROM golang:1.23-alpine3.20 AS build
WORKDIR /logporter
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o /logporter

# Final image
FROM alpine:3.20
COPY --from=build /logporter /usr/local/bin/
ENTRYPOINT ["logporter"]