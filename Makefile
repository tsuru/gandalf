get: get-test get-prod

get-test:
	@echo "Installing tests dependencies"
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' | sort | uniq | sed -e 's/\\s//g' |\
		sed -e 's/\\//g' | xargs go get >/dev/null 2>&1
	@echo "ok"

get-prod:
	@echo "Installing production dependencies"
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | grep '^.*\..*/.*$$' | grep -v 'github.com/timeredbull/gandalf' | sort | uniq | sed -e 's/\\s//g' |\
		sed -e 's/\\//g' | xargs go get >/dev/null 2>&1
	@echo "ok"

test:
	@go test -i ./...
	@go test ./...
