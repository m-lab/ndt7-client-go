FROM golang:1.18-alpine as ndt7-client-build
WORKDIR /go/src/github.com/m-lab/ndt7-client
ADD . ./
RUN go get ./cmd/ndt7-client
RUN go build ./cmd/ndt7-client

FROM alpine:3.16
WORKDIR /app
COPY --from=ndt7-client-build /go/src/github.com/m-lab/ndt7-client/ndt7-client ./
EXPOSE 8080
ENTRYPOINT ["./ndt7-client", "--port=8080"]
CMD ["--quiet", "--daemon"]
