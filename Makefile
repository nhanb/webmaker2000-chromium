build:
	go build -o dist/ .

watch:
	find . -name '*.go' -or -name '*.js' -or -name '*.html' | entr -rc go run .

clean:
	rm -rf dist
