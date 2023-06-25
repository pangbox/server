FROM docker.io/golang:1.20.5-alpine AS builder
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /server ./cmd/minibox

FROM docker.io/alpine:latest
EXPOSE 8080/tcp 10101/tcp 20202/tcp 30303/tcp
USER 1000:1000
COPY --from=builder /server /server
WORKDIR /minibox
VOLUME /minibox
ENTRYPOINT ["/server", "-pangya_dir", "/pangya"]
