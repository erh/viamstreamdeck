
BINARY = bin/viamstreamdeck

$(BINARY): bin *.go cmd/module/*.go *.mod assets/*/* assets/*/*
	go build -o $(BINARY) cmd/module/cmd.go

test:
	go test

lint:
	gofmt -w -s .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

module: $(BINARY) meta.json
	-mkdir dist # old hack	
	cp $(BINARY) dist/main
	tar czf module.tar.gz $(BINARY) meta.json dist/main

bin:
	-mkdir bin

