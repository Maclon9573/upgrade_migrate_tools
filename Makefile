ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION=${GITTAG}
GOVERSION=$(shell go version)
WORKSPACE=$(shell pwd)


export LDFLAG=-ldflags "-X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION}' \
-X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH}' \
-X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${VERSION}' \
-X 'github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME}'"

export PACKAGEPATH=./build/tools.${VERSION}

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	go mod tidy
	go fmt ./...

build:pre
	cd ${WORKSPACE} && go build ${LDFLAG} -o ${PACKAGEPATH}/cluster-migrate-tool ./main.go
