FROM golang:1.24-alpine AS builder

WORKDIR /usr/local/src

COPY ["go.mod","go.sum","./"]
RUN go mod download

COPY . .
RUN go build -o ./bin/app cmd/main/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /

EXPOSE 8080

COPY config.yaml /

CMD ["/app"]