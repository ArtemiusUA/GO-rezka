# builder image
FROM golang:1.15-alpine as builder
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
COPY templates /usr/local/share/go_rezka/templates
ENV GOREZKA_TEMPLATES_PATH=/usr/local/share/go_rezka/templates
EXPOSE 8000
