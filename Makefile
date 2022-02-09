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
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/Songmu/ghch/cmd/ghch@v0.10.2
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.2.0
	go install github.com/securego/gosec/v2/cmd/gosec@v2.9.6

prerelease:
	git pull origin --tag
	ghch -w -N ${VER}
	gocredits . > CREDITS
	git add CHANGELOG.md CREDITS
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	git push origin --tag
	goreleaser --rm-dist

.PHONY: default test
