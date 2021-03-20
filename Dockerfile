ARG REGISTRY=gcr.io/your-google-project

FROM golang:1.16-alpine as build-env

RUN apk add --no-cache build-base git
WORKDIR /build
COPY go.mod go.sum /build/
RUN go mod download
COPY . .
RUN go build -v -o /usr/local/bin/glogs-to-honeycomb

FROM $REGISTRY/base-alpine:3.1.0
COPY --from=build-env /usr/local/bin/glogs-to-honeycomb /usr/local/bin/glogs-to-honeycomb
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

EXPOSE 8080

CMD ["entrypoint.sh"]
