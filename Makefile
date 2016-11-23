# khan
# https://github.com/topfreegames/khan
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

.PHONY: db

PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/khan@.@g | egrep -v "^[.]$$")
PMD = "pmd-bin-5.3.3"
OS = "$(shell uname | awk '{ print tolower($$0) }')"
MYIP = $(shell ifconfig | egrep inet | egrep -v inet6 | egrep -v 127.0.0.1 | awk ' { print $$2 } ' | head -n 1)

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

setup: setup-ci setup-hooks
	@go get -v github.com/spf13/cobra/cobra
	@go get github.com/fzipp/gocyclo
	@go get github.com/gordonklaus/ineffassign

setup-docs:
	@pip install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

setup-ci:
	@go get github.com/mailru/easyjson/...
	@go get -u github.com/Masterminds/glide/...
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get github.com/topfreegames/goose/cmd/goose
	@go get github.com/mattn/goveralls
	@go get github.com/onsi/ginkgo/ginkgo
	@glide install

build:
	@go build $(PACKAGES)
	@go build -o ./bin/khan main.go

assets:
	@for pkg in $(GODIRS) ; do \
		go generate -x $$pkg ; \
    done

start-deps: stop-deps
	@cd ./scripts && docker-compose --project-name=khan up -d
	@until docker exec khan_postgres_1 pg_isready; do echo 'Waiting for Postgres...' && sleep 1; done
	@until docker exec khan_elasticsearch_1 curl localhost:9200; do echo 'Waiting for Elasticsearch...' && sleep 1; done
	@sleep 5
	@curl -XPUT 'http://localhost:9200/khan-test/' -d '{ "settings" : { "index" : { "number_of_shards" : 1, "number_of_replicas" : 1 } } }'
	@sleep 5
	@curl -XPUT 'http://localhost:9200/khan-test/test/1' -d '{"user" : "whatever"}'
	@sleep 5
	@docker exec khan_postgres_1 createuser -s -U postgres khan; true
	@docker exec khan_postgres_1 createdb -U khan khan; true
	@make migrate

stop-deps:
	@cd ./scripts && docker-compose --project-name=khan stop
	@cd ./scripts && docker-compose --project-name=khan rm -f

cross: assets
	@mkdir -p ./bin
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/khan-linux-i386
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/khan-linux-x86_64
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/khan-darwin-i386
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/khan-darwin-x86_64
	@chmod +x bin/*

install:
	@go install

run-verbose:
	@go run main.go start -d -c ./config/local.yaml

run:
	@go run main.go start -q -c ./config/local.yaml

run-fast:
	@go run main.go start -q --fast -c ./config/local.yaml

build-docker:
	@docker build -t khan .

build-dev-docker:
	@cp ./config/default.yaml ./dev
	@cp ./bin/khan-linux-x86_64 ./dev
	@cd dev && docker build -t khan-dev .

build-prune-docker:
	@docker build -t khan-prune -f PruneDockerfile .

# the crypto
run-docker:
	@docker run -i -t --rm \
		-e "KHAN_POSTGRES_HOST=`ifconfig | egrep inet | egrep -v inet6 | egrep -v 127.0.0.1 | awk ' { print $$2 } '`" \
		-e "KHAN_POSTGRES_PORT=5433" \
		-e "SERVER_NAME=localhost" \
		-e "AUTH_USERNAME=auth-username" \
		-e "AUTH_PASSWORD=auth-password" \
		-p 8080:80 \
		khan

run-prune-docker:
	@docker run -i -t --rm \
		-e "KHAN_POSTGRES_HOST=${MYIP}" \
		-e "KHAN_POSTGRES_PORT=5433" \
		-e "KHAN_PRUNING_SLEEP=10" \
		khan-prune

test: schema-update start-deps assets drop-test db-test
	@ginkgo -r --cover .

test-coverage coverage: test
	@echo "mode: count" > coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> coverage-all.out; done'

test-coverage-html coverage-html: test-coverage
	@go tool cover -html=coverage-all.out

random-data:
	@go run perf/main.go -games 5 -pwc 100 -cpg 10 -use-main

drop:
	@psql -d postgres -U postgres -p 5433 -h ${MYIP} -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

db migrate:
	@go run main.go migrate -c ./config/local.yaml

db-test migrate-test:
	@psql -h localhost -p 5433 -U postgres -d postgres -c "SHOW SERVER_VERSION"
	@go run main.go migrate -c ./config/test.yaml
	@go run main.go migrate -t 0 -c ./config/test.yaml
	@go run main.go migrate -c ./config/test.yaml

drop-test:
	@-psql -d postgres -h localhost -p 5433 -U postgres -c "SELECT pg_terminate_backend(pid.pid) FROM pg_stat_activity, (SELECT pid FROM pg_stat_activity where pid <> pg_backend_pid()) pid WHERE datname='khan_test';"
	@psql -d postgres -h localhost -p 5433 -U postgres -f db/drop-test.sql > /dev/null
	@echo "Test database created successfully!"

run-test-khan: build kill-test-khan
	@rm -rf /tmp/khan-bench.log
	@./bin/khan start -p 8888 -q --fast -c ./config/perf.yaml 2>&1 > /tmp/khan-bench.log &

kill-test-khan:
	@-ps aux | egrep './bin/khan.+perf.yaml' | egrep -v grep | awk ' { print $$2 } ' | xargs kill -9

ci-perf: drop-perf migrate-perf run-test-khan run-perf

run-perf:
	@go test -bench . -benchtime 3s ./bench/...

db-perf:
	@go run perf/main.go

restore-perf:
	@psql -d postgres -h localhost -p 5433 -U postgres khan_perf < khan-perf.dump

dump-perf:
	@pg_dump khan_perf > khan-perf.dump

drop-perf:
	@psql -d postgres -h localhost -p 5433 -U postgres -f db/drop-perf.sql > /dev/null
	@echo "Perf database created successfully!"

migrate-perf:
	@go run main.go migrate -c ./config/perf.yaml

static:
	@-go vet $(PACKAGES)
	@-gocyclo -over 5 . | egrep -v vendor/
	@for pkg in $$(go list ./... | grep -v /vendor/ | grep -v "/db") ; do \
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

schema-update: schema-clean
	@go generate ./models/*.go
	@go generate ./api/payload.go

schema-clean:
	@rm -rf ./models/*easyjson.go
