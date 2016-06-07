#Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)

setup:
	@brew install glide
	@go get -v github.com/spf13/cobra/cobra
	@glide install

setup-ci:
	@sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
	@sudo apt-get install -y glide
	@go get -v github.com/spf13/cobra/cobra
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

test: drop-test
	@go test $(PACKAGES)

coverage:
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)

coverage-html:
	@go tool cover -html=coverage-all.out

drop:
	@psql -d postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

drop-test:
	@psql -d postgres -f db/drop-test.sql > /dev/null
	@echo "Test database created successfully!"
