FROM golang:1.19

WORKDIR /app

COPY . .

RUN go mod init backend
RUN go build -o backend .

EXPOSE 8082

CMD ["./backend"]