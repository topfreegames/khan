#Copyright © 2016 Top Free Games <backend@tfgco.com>

PACKAGES = $(shell glide novendor)

setup:
	@brew install glide
	@go get -v github.com/spf13/cobra/cobra
	@glide install

build:
	@go build ./...
	@go build

test:
	@go test $(PACKAGES)

coverage:
	@echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	@go tool cover -html=coverage-all.out

#run:
	#@$(MAKE) rcluster &
	#@sleep 2 && forego start

#rcluster:
	#@rm -rf redis-cluster/7000/appendonly.aof redis-cluster/7000/dump.rdb redis-cluster/7000/nodes.conf
	#@rm -rf redis-cluster/7001/appendonly.aof redis-cluster/7001/dump.rdb redis-cluster/7001/nodes.conf
	#@rm -rf redis-cluster/7002/appendonly.aof redis-cluster/7002/dump.rdb redis-cluster/7002/nodes.conf
	#@rm -rf redis-cluster/7003/appendonly.aof redis-cluster/7003/dump.rdb redis-cluster/7003/nodes.conf
	#@rm -rf redis-cluster/7004/appendonly.aof redis-cluster/7004/dump.rdb redis-cluster/7004/nodes.conf
	#@rm -rf redis-cluster/7005/appendonly.aof redis-cluster/7005/dump.rdb redis-cluster/7005/nodes.conf
	#@sleep 5
	#@redis-trib.py start_multi 127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002
	#@redis-trib.py replicate 127.0.0.1:7000 127.0.0.1:7003
	#@redis-trib.py replicate 127.0.0.1:7001 127.0.0.1:7004
	#@redis-trib.py replicate 127.0.0.1:7002 127.0.0.1:7005

#bench:
	#@echo "Make sure thor is installed with 'npm install -g thor'"
	#@thor --buffer 40 --generator thor.js --amount 500 --messages 1000 --masked ws://localhost:9999/echo

#watch:
	#@react2fs --include "^.+[.]go$$" make test

#drop:
	#@psql -d postgres -f db/drop.sql > /dev/null
	#@echo "Database created successfully!"

#drop-test:
	#@psql -d postgres -f db/drop-test.sql > /dev/null
	#@echo "Test database created successfully!"

#update-schema:
	#@./flatc --go schema.fbs --binary
	#@rm -f ./channel/schema.js
	#@./flatc --js schema.fbs --binary 
	#@mv schema_generated.js ./channel/schema.js
	#@rm -rf models/*ffjson*
	#@go generate ./...
