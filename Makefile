#Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/khan@.@g | egrep -v "^[.]$$")

setup:
	@brew install glide
	@go get -v github.com/spf13/cobra/cobra
	@go get github.com/fzipp/gocyclo
	@go get bitbucket.org/liamstask/goose/cmd/goose
	@go get github.com/fzipp/gocyclo
	@go get github.com/gordonklaus/ineffassign
	@glide install

setup-ci:
	@sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
	@sudo apt-get install -y glide
	@go get github.com/mattn/goveralls
	@glide install

build:
	@go build $(PACKAGES)
	@go build

run:
	@go run main.go start -d -c ./config/local.yaml

build-docker:
	@docker build -t khan .

run-docker:
	@docker run -i -t --rm -e "KHAN_POSTGRES_HOST=10.0.20.81" -p 8080:8080 khan

test: drop-test db-test
	@go test $(PACKAGES)

coverage:
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html:
	@go tool cover -html=coverage-all.out

db migrate:
	@goose -env development up

drop:
	@psql -d postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

db-test migrate-test:
	@goose -env test up

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

