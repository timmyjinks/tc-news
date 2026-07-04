FROM golang:1.26

WORKDIR /app

COPY go.sum go.mod .
RUN go mod download

COPY . .
RUN go build -o /bin/app ./cmd

CMD ["/bin/app"]
