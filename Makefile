watch:
	find . -name '*.go' | entr -rc go run .

compile:
	go build -o dist/ .

clean:
	rm -rf dist
