BUILD_VERSION   	:= $(shell cat version)
BUILD_DATE      	:= $(shell date "+%F %T")
COMMIT_SHA1     	:= $(shell git rev-parse HEAD)
DOCKER_IMAGE 	    := mritd/goadmission

all: clean
	gox -osarch="darwin/amd64 linux/386 linux/amd64 linux/arm" \
		-output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}" \
		-ldflags	"-X 'main.version=${BUILD_VERSION}' \
					-X 'main.buildDate=${BUILD_DATE}' \
					-X 'main.commitID=${COMMIT_SHA1}'"

release: all
	tar -zcf dist/deploy.tar.gz deploy
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate -name "Bump ${BUILD_VERSION}" --debug ${BUILD_VERSION} dist

install:
	go install -ldflags	"-X 'main.version=${BUILD_VERSION}' \
               			-X 'main.buildDate=${BUILD_DATE}' \
               			-X 'main.commitID=${COMMIT_SHA1}'"

docker:
	cat Dockerfile | docker build --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} -t ${DOCKER_IMAGE}:${BUILD_VERSION} -f - .
	docker tag ${DOCKER_IMAGE}:${BUILD_VERSION} ${DOCKER_IMAGE}:latest

docker-push: docker
	docker push ${DOCKER_IMAGE}:${BUILD_VERSION}
	docker push ${DOCKER_IMAGE}:latest

clean:
	rm -rf dist

.PHONY: all release clean install docker

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
