FROM golang:1.10-alpine as builder
RUN apk add --update make git
WORKDIR src/git.containerum.net/ch/kube-api
COPY . .
RUN VERSION=$(git describe --abbrev=0 --tags) make build-for-docker

FROM alpine:3.7

VOLUME ["/cfg"]

COPY --from=builder /tmp/kube-api /
ENV CH_KUBE_API_DEBUG="true" \
    CH_KUBE_API_TEXTLOG="true"

EXPOSE 1212

CMD ["/kube-api"]