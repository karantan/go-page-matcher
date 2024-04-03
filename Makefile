.PHONY: build clean deploy

build:
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bootstrap functions/page_matcher/main.go

clean:
	rm -rf ./bin bootstrap

deploy: clean build
	sls deploy --verbose --aws-profile <AWS-PROFILE>
