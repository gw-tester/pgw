FROM golang:1.16-alpine3.13 as build

WORKDIR /go/src/github.com/electrocucaracha/pgw

ENV GO111MODULE "on"
ENV CGO_ENABLED "0"
ENV GOOS "linux"
ENV GOARCH "amd64"
ENV GOBIN=/bin

COPY go.mod go.sum ./
COPY ./internal/imports ./internal/imports
RUN go build ./internal/imports
COPY . .
RUN go build -v -o /bin ./...

FROM build as test
RUN go test -v ./...

FROM alpine:3.13

ENV S5U_NETWORK "172.25.0.0/24"
ENV S5C_NETWORK "172.25.1.0/24"
ENV SGI_NIC "eth2"
ENV SGI_SUBNET "10.0.1.0/24"
ENV REDIS_URL ""
ENV LOG_LEVEL ""

COPY --from=build /bin/cmd /pwg

RUN apk add --no-cache tini=0.19.0-r0
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/pwg"]
