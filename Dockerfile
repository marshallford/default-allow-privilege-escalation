FROM golang:1.15.1-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

WORKDIR /go/src/app/cmd/webhook
RUN go build -o /go/bin/app

FROM gcr.io/distroless/base-debian10:nonroot

ARG MAINTAINER
ARG CREATED
ARG REVISION
ARG VERSION
ARG TITLE
ARG REPOSITORY_URL

LABEL maintainer=$MAINTAINER
LABEL org.opencontainers.image.created=$CREATED \
      org.opencontainers.image.revision=$REVISION \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.title=$TITLE \
      org.opencontainers.image.source=$REPOSITORY_URL \
      org.opencontainers.image.url=$REPOSITORY_URL

COPY --from=build /go/bin/app /
CMD ["/app"]
EXPOSE 8443
