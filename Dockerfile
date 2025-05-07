FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download
RUN go mod tidy

COPY . ./

RUN go test ./... -cover

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /resizer ./cmd/resizer/

FROM scratch

COPY --from=builder /resizer /resizer

ENTRYPOINT ["/resizer"]
