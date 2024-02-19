ARG GO_VERSION=1.21.6

FROM golang:${GO_VERSION}-alpine AS builder

RUN go env -w GOPROXY=direct
# --no-cache es para asegurarse que se descargue
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates && update-ca-certificates

WORKDIR /src

COPY ./go.mod ./go.sum ./

# instalamos todas las dependencias definidas en el archivo .mod
RUN go mod download

COPY events events
COPY repository repository
COPY database database
COPY search search
COPY models models
COPY feedService feedService

RUN go install ./...

# ****************************** NEW SERVICE ****************************** # 
FROM alpine:3.11

WORKDIR /usr/bin

COPY --from=builder /go/bin .