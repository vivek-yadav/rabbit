FROM golang:1.16-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Unit tests
RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/app .

# Start fresh from a smaller image
FROM alpine
RUN apk add ca-certificates

WORKDIR /app

COPY --from=build_base /tmp/app/out/app /app/app
COPY --from=build_base /tmp/app/.rabbit.yaml /app/.rabbit.yaml

EXPOSE 9099

CMD [ "/app/app", "serve" ]