FROM golang:1.25

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go .
COPY internal/ ./internal/
COPY web/ ./web/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main

CMD ["/main"]