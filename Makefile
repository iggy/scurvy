GO_FILES    := $(shell find . -iname '*.go' -type f)
DATETIME    := $(shell date +%Y%m%d%H%M)

# This stage can be run locally to install tools to host system
local_prep:
	go get github.com/golang/lint/golint
	go get honnef.co/go/tools/cmd/megacheck
	go get github.com/fzipp/gocyclo
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

host_check: local_prep
	gofmt -l -s -w ./irc ./cmd ./pkg
	test -z $(gofmt -s -l $GO_FILES)
	go test -v -race ./...
	go vet ./...
	megacheck ./...
	gocyclo -over 19 $GO_FILES
	golint -set_exit_status $(go list ./...)

host_build: host_check
	# build the github release files
	- gox -arch="amd64 arm 386" -os="linux" -output="dist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...

	# build the Docker release (this has CGO_ENABLED=0 for static binary building
	# for use in scratch images)
	- mkdir -p $TRAVIS_BUILD_DIR/ddist/etc
	- CGO_ENABLED=0 gox -arch="amd64" -os="linux" -output="ddist/{{.OS}}_{{.Arch}}_{{.Dir}}" -ldflags='-extldflags "-static" -s -w' -tags='netgo' ./...
	- $TRAVIS_BUILD_DIR/upx $TRAVIS_BUILD_DIR/ddist/linux*

# This is really just an optimization so we aren't downloading/installing
# the same packages over and over
# Also lets us use the same method to build locally and in travis
docker:
	docker build --build-arg=BUILDPLATFORM=linux/amd64 --pull --tag scurvy:test .
	docker build --build-arg=BUILDPLATFORM=linux/amd64 --target build --tag scurvy:build .
	docker build --build-arg=BUILDPLATFORM=linux/amd64 --target irc --tag notiggy/scurvy-irc .
	docker build --build-arg=BUILDPLATFORM=linux/amd64 --target notifyd --tag notiggy/scurvy-notifyd .
	docker build --build-arg=BUILDPLATFORM=linux/amd64 --target input-webhook --tag notiggy/scurvy-input-webhook .
	# the below line needs a _very_ new version of docker (i.e. experimental)
	# I will just run this locally for now, but need to hook it up to CI later
	# docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t notiggy/scurvy-irc:latest --target irc --pull --push .
	# docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t notiggy/scurvy-notifyd:latest --target notifyd --pull --push .
	# docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t notiggy/scurvy-input-webhook:latest --target input-webhook --pull --push .
	# docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t notiggy/scurvy-syncd:latest --target syncd --pull --push .

release:
	# do the github release
	docker run -e GITHUB_TOKEN scurvy:build ghr --repository scurvy --username iggy --replace $(shell date +%Y%m%d%H%M) dist/
	# do the docker hub release
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)
	docker push notiggy/scurvy-irc
	docker push notiggy/scurvy-notifyd
	docker push notiggy/scurvy-input-webhook
