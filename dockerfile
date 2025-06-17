# Build docker cli
FROM golang:1.23-alpine3.20 AS build
RUN apk add -U -q --progress --no-cache git bash coreutils gcc musl-dev
WORKDIR /go/src/github.com/docker/cli
RUN git clone --branch v27.0.3 --single-branch --depth 1 https://github.com/docker/cli .
ENV CGO_ENABLED=0
ENV GOARCH=amd64
ENV DISABLE_WARN_OUTSIDE_CONTAINER=1
RUN ./scripts/build/binary
RUN rm build/docker && mv build/docker-linux-* build/docker

# Build logporter
WORKDIR /logporter
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o /logporter

# Final image
FROM alpine:3.20
COPY --from=build /go/src/github.com/docker/cli/build/docker /bin/docker
COPY --from=build /logporter /usr/local/bin/
ENTRYPOINT ["logporter"]