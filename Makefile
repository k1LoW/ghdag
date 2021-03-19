PKG = github.com/k1LoW/ghdag
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

export GO111MODULE=on

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

default: test

ci: depsdev test integration sec

test:
	mockgen -source gh/gh.go -destination mock/mock_gh.go -package mock
	mockgen -source slk/slk.go -destination mock/mock_slk.go -package mock
	go test ./... -coverprofile=coverage.txt -covermode=count

integration: build
	./ghdag run testdata/test_workflow.yml

sec:
	gosec ./...

build:
	go build -ldflags="$(BUILD_LDFLAGS)"

depsdev:
	go get github.com/golang/mock/mockgen
	go get github.com/Songmu/ghch/cmd/ghch
	go get github.com/Songmu/gocredits/cmd/gocredits
	go get github.com/securego/gosec/cmd/gosec

prerelease:
	git pull origin --tag
	ghch -w -N ${VER}
	gocredits . > CREDITS
	git add CHANGELOG.md CREDITS
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	goreleaser --rm-dist

.PHONY: default test
