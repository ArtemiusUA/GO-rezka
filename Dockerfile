# builder image
FROM golang:alpine as builder
RUN mkdir /src
RUN mkdir /build
COPY cmd /src/cmd
COPY internal /src/internal
COPY go.mod /src
COPY go.sum /src
WORKDIR /src
RUN go mod download
WORKDIR /src/cmd/collector/
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /build/collector .
WORKDIR /src/cmd/web
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /build/web .
RUN ls /build

# final image
FROM alpine
COPY --from=builder /build/collector /usr/local/bin/
COPY --from=builder /build/web /usr/local/bin/
