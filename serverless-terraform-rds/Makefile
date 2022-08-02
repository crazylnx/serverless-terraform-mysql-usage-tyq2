# 修改这里
VERSION ?= 0.0.1
# BUILD_TIME is used for daemon BuildTime in go build.
BUILD_TIME=$(shell date --rfc-3339 s 2> /dev/null | sed -e 's/ /T/')

# 修改这里
IMG = registry.cn-hangzhou.aliyuncs.com/fc-demo2/serverless-terraform-mysql-creator

fmt:
	go fmt ./...

build-img:
	docker build -t ${IMG}:${VERSION} -f Dockerfile .

run:
	docker run -p 9000:9000 --name myTerraform --rm -it ${IMG}:${VERSION}

push:
	docker push ${IMG}:${VERSION}

vendor:
	go mod vendor


build&push: build-img push
