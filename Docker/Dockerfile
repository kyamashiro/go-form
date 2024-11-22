FROM golang:1.23

WORKDIR /app

RUN go install github.com/air-verse/air@latest

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

CMD ["air", "-c", ".air.toml"]