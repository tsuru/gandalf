get-test:
	@/bin/echo -n "Installing test dependencies... "
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' | sort | uniq | sed -e 's/\\s//g' |\
		sed -e 's/\\//g' | xargs go get -u -v
	@go list -f '{{range .XTestImports}}{{.}}\
		{{end}}' ./... | grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' | sort | uniq | sed -e 's/\\s//g' |\
		sed -e 's/\\//g' | xargs go get -u -v
	@/bin/echo "ok"

test:
	@go test -i ./...
	@go test ./...
