FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

COPY vendor vendor

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -mod=vendor -o 3s .


FROM scratch

COPY --from=builder /app/3s /3s

ENTRYPOINT ["/3s"]
