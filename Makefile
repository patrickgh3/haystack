all: haystack

haystack: *
	go install
	cp $$GOPATH/bin/haystack ~/haystack-test/
	cp -r $$GOPATH/src/github.com/patrickgh3/haystack/templates ~/haystack-test
	cp -r $$GOPATH/src/github.com/patrickgh3/haystack/assets/* /var/html/cwpat.me/haystack-dev

run: haystack
	~/haystack-test/haystack

