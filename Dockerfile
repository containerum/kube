FROM golang:1.9-alpine as builder
WORKDIR /go/src/git.containerum.net/ch/kube-api
COPY . .
WORKDIR cmd/kube-api
RUN CGO_ENABLED=0 go build -v -ldflags="-w -s -extldflags '-static'" -tags="jsoniter" -o /bin/kube-api

FROM alpine:3.7
COPY --from=builder /bin/kube-api /
ENV CH_KUBE_API_KUBE_CONF="/cfg/kube.conf" \
    CH_KUBE_API_DEBUG="true" \
    CH_KUBE_API_TEXTLOG="true"
VOLUME ["/cfg"]
EXPOSE 1212
ENTRYPOINT ["/kube-api"]
