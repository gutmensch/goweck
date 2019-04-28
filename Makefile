goweck : clean deps bindata lint test build

deps :
	@echo === Collecting dependencies
	-go get -u github.com/go-bindata/go-bindata/...
	-go get -u github.com/globalsign/mgo
	-go get -u github.com/globalsign/mgo/bson
	-go get -u github.com/gorilla/mux
	-go get -u github.com/gregdel/pushover
	-go get -u github.com/imdario/mergo

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

run :
	@echo === Running GoWeck for Testing
	MONGODB_URI=mongodb://192.168.1.100:27017 \
	MONGODB_DATABASE=goweck_test \
	MONGODB_DROP=true \
	RAUMSERVER_URI=http://qnap:3535/raumserver \
	PUSHOVER_APP_TOKEN=anar1e74bu8ro1qnaf4oky61a7aakf \
	PUSHOVER_USER_TOKEN=uq5iMscx8WzoBMKiCEJMRhrgaEhYyE \
	DEEP_STANDBY=false \
	TZ=Europe/Berlin \
	LISTEN=:8080 \
	DEBUG=true \
	./goweck
