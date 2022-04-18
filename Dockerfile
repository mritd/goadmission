FROM golang:alpine AS builder

ENV SRC_PATH ${GOPATH}/src/github.com/mritd/goadmission

WORKDIR ${SRC_PATH}

COPY . .

RUN set -ex \
    && apk add git \
    && export BUILD_VERSION=$(cat version) \
    && export BUILD_DATE=$(date "+%F %T") \
    && export COMMIT_SHA1=$(git rev-parse HEAD) \
    && go install -trimpath -ldflags \
        "-X 'main.version=${BUILD_VERSION}' \
        -X 'main.buildDate=${BUILD_DATE}' \
        -X 'main.commitID=${COMMIT_SHA1}' \
        -w -s"

FROM alpine

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8
ENV LANGUAGE en_US:en

RUN set -ex \
    && apk add bash tzdata ca-certificates \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/goadmission /goadmission

ENTRYPOINT ["/goadmission"]

CMD ["--help"]
