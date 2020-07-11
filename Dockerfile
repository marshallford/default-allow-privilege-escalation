FROM golang:1.14-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

WORKDIR /go/src/app/cmd/webhook
RUN go build -o /go/bin/app

FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/app /
CMD ["/app"]
EXPOSE 8443
