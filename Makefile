BUILD_DIR = build
GANDALF_WEBSERVER_BIN = $(BUILD_DIR)/gandalf-webserver
GANDALF_WEBSERVER_SRC = webserver/main.go
GANDALF_SSH_BIN = $(BUILD_DIR)/gandalf-ssh
GANDALF_SSH_SRC = bin/gandalf.go

get: get-code godep binaries

get-code:
	go get $(GO_EXTRAFLAGS) -u -d -t ./...

godep:
	go get $(GO_EXTRAFLAGS) github.com/tools/godep
	godep restore ./...

test: get-code
	go clean $(GO_EXTRAFLAGS) ./...
	go test $(GO_EXTRAFLAGS) ./...

doc:
	@cd docs && make html

binaries: gandalf-webserver gandalf-ssh

gandalf-webserver: $(GANDALF_WEBSERVER_BIN)

$(GANDALF_WEBSERVER_BIN):
	godep go build -o $(GANDALF_WEBSERVER_BIN) $(GANDALF_WEBSERVER_SRC)

run-gandalf-webserver: $(GANDALF_WEBSERVER_BIN)
	$(GANDALF_WEBSERVER_BIN) $(GANDALF_WEBSERVER_OPTIONS)

gandalf-ssh: $(GANDALF_SSH_BIN)

$(GANDALF_SSH_BIN):
	godep go build -o $(GANDALF_SSH_BIN) $(GANDALF_SSH_SRC)

run-gandalf-ssh: $(GANDALF_SSH_BIN)
	$(GANDALF_SSH_BIN) $(GANDALF_SSH_OPTIONS)
