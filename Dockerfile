FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download
RUN go mod tidy

COPY . ./

# RUN go test ./internal/... -cover

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/resizer ./cmd/resizer
# Сборка бинарника с отладочной информацией
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#    go build -gcflags="all=-N -l" -o /bin/resizer ./cmd/resizer

FROM scratch

COPY --from=builder /bin/resizer /resizer

ENTRYPOINT ["/resizer"]