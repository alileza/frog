project_name = frog
branch = $(shell git symbolic-ref HEAD 2>/dev/null)
version = 0.1.0
revision = $(shell git log -1 --pretty=format:"%H")
build_user = $(USER)
build_date = $(shell date +%FT%T%Z)
pwd = $(shell pwd)

build_dir ?= bin/

pkgs          = ./...
version_pkg= github.com/alileza/frog/util/version
ldflags := "-X $(version_pkg).Version=$(version) -X $(version_pkg).Branch=$(branch) -X $(version_pkg).Revision=$(revision) -X $(version_pkg).BuildUser=$(build_user) -X $(version_pkg).BuildDate=$(build_date)"


deps:
	@echo " > Installing dependencies"
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure --vendor-only

build:
	@echo ">> building binaries"
	@go build -ldflags $(ldflags) -o $(build_dir)/$(project_name) main.go

build-all:
	@echo ">> packaging releases"
	@rm -rf dist
	@mkdir dist
	@for os in "linux" "darwin" ; do \
			for arch in "amd64" "386" "arm" "arm64" ; do \
					echo " > building $$os/$$arch" ; \
					GOOS=$$os GOARCH=$$arch go build -ldflags $(ldflags) -o $(build_dir)/$(project_name).$(version).$$os-$$arch main.go ; \
			done ; \
	done

package-releases:
	@echo ">> packaging releases"
	@rm -rf dist
	@mkdir dist
	@for f in $(shell ls bin) ; do \
			cp bin/$$f frog ; \
			tar -czvf $$f.tar.gz frog ; \
			mv $$f.tar.gz dist ; \
			rm -rf frog ; \
	done
