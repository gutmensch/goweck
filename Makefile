goweck : clean bindata lint test build

bindata :
	@echo === Converting static assets to bindata
	go-bindata asset/

lint :
	@echo === Running style checks
	golint

test :
	@echo === Running tests
	go test

build :
	@echo === Building goweck
	go build

clean :
	@echo === Cleaning up
	-rm -f goweck
