# builder image
FROM golang:alpine as builder
RUN mkdir /src
RUN mkdir /build
COPY . /src/
WORKDIR /src
RUN go mod download
WORKDIR /src/cmd/collector
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /build/collector .
WORKDIR /src/cmd/web
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /build/web .

# final image
FROM alpine:latest
COPY --from=builder /build/collector /usr/local/bin/
COPY --from=builder /build/web /usr/local/bin/
