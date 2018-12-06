all: prepare fetch-dep bld

prepare: 
	echo "run all"
	mkdir -p ./build/_workspace


bld:
	echo "start build"
	GOPATH=`pwd`/build/_workspace go build -v -o ./istanbulextradatagen ./main.go
	@echo "Done building."


fetch-dep:
	GOPATH=`pwd`/build/_workspace go get -u github.com/ethereum/go-ethereum
	GOPATH=`pwd`/build/_workspace go get -u golang.org/x/crypto/ssh/terminal


clean:
	rm istanbulextradatagen

test:
	test "`cat addr.json | ./istanbulextradatagen`" = "Extradata for addrs: {0x0000000000000000000000000000000000000000000000000000000000000000f89af8549409bee72809bf4c99fa088cdee5d8a6ab9cb5e1b594b7b42d18a4ca0339ff364f4768b7ffdd01bb6c2794b8c7f3caf6b38e9a89c5b4902492adbc43fb3b9994c18ad7f5b128aeb7c689900527aca1b7d7069bb9b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0}"