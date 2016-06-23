# khan
# https://github.com/topfreegames/khan
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/khan@.@g | egrep -v "^[.]$$")
PMD = "pmd-bin-5.3.3"

setup:
	@go get -u github.com/Masterminds/glide/...
	@go get -v github.com/spf13/cobra/cobra
	@go get github.com/fzipp/gocyclo
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/fzipp/gocyclo
	@go get github.com/gordonklaus/ineffassign
	@go get -u github.com/jteeuwen/go-bindata/...
	@glide install

setup-docs:
	@pip install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

setup-ci:
	@sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
	@sudo apt-get install -y glide
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/mattn/goveralls
	@glide install

build:
	@go build $(PACKAGES)
	@go build

assets:
	@for pkg in $(GODIRS) ; do \
		go generate -x $$pkg ; \
    done

cross: assets
	@mkdir -p ./bin
	@echo "Building for linux-386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/khan-linux-386
	@echo "Building for linux-amd64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/khan-linux-amd64
	@echo "Building for darwin-386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/khan-darwin-386
	@echo "Building for darwin-amd64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/khan-darwin-amd64

install:
	@go install

run:
	@go run main.go start -d -v3 -c ./config/local.yaml

build-docker:
	@docker build -t khan .

run-docker:
	@docker run -i -t --rm -e "KHAN_POSTGRES_HOST=10.0.20.81" -p 8080:8080 khan

test: assets drop-test db-test
	@go test $(PACKAGES)

coverage: drop-test db-test
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg) && \
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html:
	@go tool cover -html=coverage-all.out

db migrate:
	@go run main.go migrate -c ./config/local.yaml

drop:
	@psql -d postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

db-test migrate-test:
	@psql -d postgres -c "SHOW SERVER_VERSION"
	@go run main.go migrate -c ./config/test.yaml

drop-test:
	@psql -d postgres -f db/drop-test.sql > /dev/null
	@echo "Test database created successfully!"

static:
	@go vet $(PACKAGES)
	@-gocyclo -over 5 . | egrep -v vendor/
	@#golint
	@for pkg in $$(go list ./... |grep -v /vendor/) ; do \
        golint $$pkg ; \
    done
	@#ineffassign
	@for pkg in $(GODIRS) ; do \
        ineffassign $$pkg ; \
    done
	@${MAKE} pmd

pmd:
	@bash pmd.sh
	@for pkg in $(GODIRS) ; do \
		exclude=$$(find $$pkg -name '*_test.go') && \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --exclude $$exclude --language go ; \
    done

pmd-full:
	@bash pmd.sh
	@for pkg in $(GODIRS) ; do \
		/tmp/pmd-bin-5.4.2/bin/run.sh cpd --minimum-tokens 30 --files $$pkg --language go ; \
    done

rtfd:
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html
