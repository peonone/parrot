.PHONY: build protogen test 

build:
	./build.sh

protogen:
	for d in auth user chat; do \
	    protoc --proto_path=. --micro_out=. --go_out=. $${d}/proto/*.proto;\
	done

test:
	go test -v ./...
