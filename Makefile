watch:
	find . -name '*.go' -or -name '*.js' -or -name '*.html' | entr -rc go run .

compile:
	go build -o dist/ .

clean:
	rm -rf dist
