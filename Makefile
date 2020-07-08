BUILD_VERSION   	:= $(shell cat version)
BUILD_DATE      	:= $(shell date "+%F %T")
COMMIT_SHA1     	:= $(shell git rev-parse HEAD)
CSI_DOCKER_IMAGE 	:= gozap/csi-nfs

all: clean
	gox -osarch="darwin/amd64 linux/386 linux/amd64 linux/arm" \
		-output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
		-ldflags	"-X 'github.com/gozap/csi-nfs/cmd.Version=${BUILD_VERSION}' \
					-X 'github.com/gozap/csi-nfs/cmd.BuildDate=${BUILD_DATE}' \
					-X 'github.com/gozap/csi-nfs/cmd.CommitID=${COMMIT_SHA1}'"

release: docker-push clean
	mkdir dist && tar -zcf dist/deploy.tar.gz deploy
	ghr -u gozap -t ${GITHUB_TOKEN} -replace -recreate -name "Bump ${BUILD_VERSION}" --debug ${BUILD_VERSION} dist/deploy.tar.gz

install:
	go install -ldflags	"-X 'github.com/gozap/csi-nfs/cmd.Version=${BUILD_VERSION}' \
               			-X 'github.com/gozap/csi-nfs/cmd.BuildDate=${BUILD_DATE}' \
               			-X 'github.com/gozap/csi-nfs/cmd.CommitID=${COMMIT_SHA1}'"

docker:
	cat Dockerfile | docker build -t ${CSI_DOCKER_IMAGE}:${BUILD_VERSION} -f - .
	docker tag ${CSI_DOCKER_IMAGE}:${BUILD_VERSION} ${CSI_DOCKER_IMAGE}:latest

docker-push: docker
	docker push ${CSI_DOCKER_IMAGE}:${BUILD_VERSION}
	docker push ${CSI_DOCKER_IMAGE}:latest

docker-debug:
	cat Dockerfile.debug | docker build -t ${CSI_DOCKER_IMAGE}:debug -f - .

clean:
	rm -rf dist

.PHONY: all release clean install docker docker-debug

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.cn
