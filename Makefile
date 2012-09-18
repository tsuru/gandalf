get: get-test get-prod

get-test:
	@/bin/echo "Installing test dependencies... "
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | tr ' ' '\n' |\
		grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' |\
		sort | uniq | xargs go get -u -v
	@go list -f '{{range .XTestImports}}{{.}} {{end}}' ./... | tr ' ' '\n' |\
		grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' |\
		sort | uniq | xargs go get -u -v
	@/bin/echo "ok"

get-prod:
	@/bin/echo "Installing production dependencies... "
	@go list -f '{{range .Imports}}{{.}} {{end}}' ./... | tr ' ' '\n' |\
		grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' |\
		sort | uniq | xargs go get -u -v
	@/bin/echo "ok"

test:
	@go test -i ./...
	@go test ./...
