get: get-test get-prod godep

get-test:
	@/bin/echo "Installing test dependencies... "
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | tr ' ' '\n' |\
		grep '^.*\..*/.*$$' | grep -v 'github.com/tsuru/gandalf' |\
		sort | uniq | xargs go get -u >/dev/null 2>&1
	@/bin/echo "ok"

get-prod:
	@/bin/echo "Installing production dependencies... "
	@go list -f '{{range .Imports}}{{.}} {{end}}' ./... | tr ' ' '\n' |\
		grep '^.*\..*/.*$$' | grep -v 'github.com/tsuru/gandalf' |\
		sort | uniq | xargs go get -u >/dev/null 2>&1
	@/bin/echo "ok"

godep:
	go get github.com/tools/godep
	godep restore ./...
	godep go clean ./...

test:
	@go test -i ./...
	@go test ./...

doc:
	@cd docs && make html
