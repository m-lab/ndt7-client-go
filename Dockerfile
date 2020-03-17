# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:alpine as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

RUN apk update \
    && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

RUN adduser -D -g '' appuser

ADD . ${GOPATH}/src/app/
WORKDIR ${GOPATH}/src/app

RUN go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/mlab_exporter

# --------------------------------------------------------------------------------

FROM gcr.io/distroless/base

LABEL summary="mlab Speedtest Prometheus exporter" \
      description="A Prometheus exporter for broadband speedtests using m-lab" \
      name="wlbr/mlab-exporter" \
      url="https://github.com/wlbr/ndt7-client-go" \
      maintainer="Michael Wolber <mwolber@gmx.de>"

COPY --from=builder /go/bin/mlab_exporter /usr/bin/mlab_exporter

COPY --from=builder /etc/passwd /etc/passwd

EXPOSE 9122

ENTRYPOINT [ "/usr/bin/mlab_exporter", "-format", "prometheus" ]
