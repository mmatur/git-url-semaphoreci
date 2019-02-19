FROM golang:alpine3.8 as builder

RUN apk --update upgrade \
&& apk --no-cache --no-progress add make git gcc musl-dev \
&& rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/mmatur/git-url-semaphoreci
COPY . .
RUN make build

FROM alpine:3.8
RUN apk update && apk add --no-cache --virtual ca-certificates
COPY --from=builder /go/src/github.com/mmatur/git-url-semaphoreci/git-url-semaphoreci /usr/bin/git-url-semaphoreci

LABEL "name"="Git URL semaphoreci"
LABEL "com.github.actions.name"="Git URL semaphoreci"
LABEL "com.github.actions.description"="This is an Action to get git url in a semaphoreci build."
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="green"

LABEL "repository"="http://github.com/mmatur/git-url-semaphoreci"
LABEL "homepage"="http://github.com/mmatur/git-url-semaphoreci"
LABEL "maintainer"="Michael Matur <michael.matur@gmail.com>"

ENTRYPOINT [ "/usr/bin/git-url-semaphoreci" ]