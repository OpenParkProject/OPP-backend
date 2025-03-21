FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev curl
WORKDIR /go/src/app
COPY src/go.mod src/go.sum ./
RUN go mod download
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
RUN git clone -b main --depth 1 https://github.com/OpenParkProject/OPP-common.git /tmp/OPP-common
RUN mkdir -p api
RUN oapi-codegen -generate types,gin-server -o api/api.gen.go -package api /tmp/OPP-common/openapi.yaml
COPY src/ .
RUN go build -o /go/bin/opp-backend

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=0 /go/bin/opp-backend .
EXPOSE 10020

CMD ["./opp-backend"]