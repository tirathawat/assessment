FROM golang:1.19-alpine as build-base

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v --tags=unit ./...

RUN go build -o ./out/server .

# ====================


FROM alpine:3.16.2
COPY --from=build-base /app/out/server /app/server

CMD ["/app/server"]
