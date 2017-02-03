all: haystack

haystack: *.go
	./inst.sh

run: haystack
	test/haystack

clean:
	rm test/haystack

