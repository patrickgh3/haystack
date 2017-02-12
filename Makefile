all: haystack

haystack: *
	go install
	cp $$GOPATH/bin/haystack ~/haystack-test/
	cp -r $$GOPATH/src/github.com/patrickgh3/haystack/html/ ~/haystack-test/html/

run: haystack
	~/haystack-test/haystack

