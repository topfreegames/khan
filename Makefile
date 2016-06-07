test:
	@go test $$(glide novendor)

drop:
	@psql -d postgres -f db/drop.sql > /dev/null
	@echo "Database created successfully!"

drop-test:
	@psql -d postgres -f db/drop-test.sql > /dev/null
	@echo "Test database created successfully!"
