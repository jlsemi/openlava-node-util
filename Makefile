export NODECLI_VERSION=0.3.2
export TESTRUN_VERSION=0.1.0

build-nodecli:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/nodecli-${NODECLI_VERSION} ./bin/nodecli 
	chmod +x build/nodecli-${NODECLI_VERSION}

build-testrun:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/testrun-${TESTRUN_VERSION} ./bin/testrun
	chmod +x build/testrun-${TESTRUN_VERSION}
