goweck : clean deps bindata lint test build

deps :
	@echo === Collecting dependencies
	-go get -u github.com/go-bindata/go-bindata/...
	-go get -u github.com/globalsign/mgo
	-go get -u github.com/globalsign/mgo/bson
	-go get -u github.com/gorilla/mux
	-go get -u github.com/gregdel/pushover
	-go get -u github.com/imdario/mergo
	-go get -u golang.org/x/lint/golint

bindata :
	@echo === Converting static assets to bindata
	go-bindata asset/

lint :
	@echo === Running style checks
	-golint

test :
	@echo === Running tests
	go test

build :
	@echo === Building goweck
	go build -ldflags="-s -w"

clean :
	@echo === Cleaning up
	-rm -f goweck

run :
	@echo === Running GoWeck for Testing
	MONGODB_URI=mongodb://192.168.1.100:27017 \
	MONGODB_DATABASE=goweck_test \
	MONGODB_DROP=true \
	RAUMSERVER_URI=http://qnaps:3535/raumserver \
	PUSHOVER_APP_TOKEN=xxx \
	PUSHOVER_USER_TOKEN=yyy \
	DEEP_STANDBY=false \
	TZ=Europe/Berlin \
	LISTEN=:8080 \
	DEBUG=true \
	./goweck
