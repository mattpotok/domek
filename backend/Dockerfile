# syntax=docker/dockerfile:1

# Build application from source
FROM golang:latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /domek

# Deploy application to lean image
FROM gcr.io/distroless/base-debian12 AS release-stage

WORKDIR /

COPY --from=build-stage /domek /domek

EXPOSE 3333

ENTRYPOINT [ "/domek" ]
