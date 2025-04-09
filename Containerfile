FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git make gcc musl-dev curl
WORKDIR /go/src/app
COPY src/go.mod src/go.sum ./
RUN go mod download
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
RUN git clone -b main --depth 1 https://github.com/OpenParkProject/OPP-common.git /tmp/OPP-common
RUN mkdir -p api
RUN cp /tmp/OPP-common/openapi.yaml api/openapi.yaml
COPY src/ .
RUN go generate
RUN go build -o /go/bin/opp-backend

FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=0 /go/bin/opp-backend .
COPY --from=0 /go/src/app/api/openapi.yaml ./api/openapi.yaml
EXPOSE 8080

CMD ["./opp-backend"]
