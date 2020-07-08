FROM golang:1.14.4-alpine3.12 AS builder

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn
ENV SRC_PATH ${GOPATH}/src/github.com/gozap/csi-nfs

WORKDIR ${SRC_PATH}

COPY . .

RUN set -ex \
    && apk add git tar \
    && export BUILD_VERSION=$(cat version) \
    && export BUILD_DATE=$(date "+%F %T") \
    && export COMMIT_SHA1=$(git rev-parse HEAD) \
    && go install -ldflags \
        "-X 'github.com/gozap/csi-nfs/cmd.Version=${BUILD_VERSION}' \
        -X 'github.com/gozap/csi-nfs/cmd.BuildDate=${BUILD_DATE}' \
        -X 'github.com/gozap/csi-nfs/cmd.CommitID=${COMMIT_SHA1}'"

FROM alpine:3.12

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8
ENV LANGUAGE en_US:en

RUN set -ex \
    && apk add bash tzdata ca-certificates nfs-utils \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/csi-nfs /csi-nfs

ENTRYPOINT ["/csi-nfs"]

CMD ["--help"]
