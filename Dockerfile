FROM golang:1.18-alpine

WORKDIR /app
COPY . ./

RUN go get ./cmd/ndt7-client
RUN go build ./cmd/ndt7-client

EXPOSE 8080
CMD ["./ndt7-client", "--quiet", "--daemon", "--port=8080"]