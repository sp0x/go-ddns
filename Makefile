NAME=ddns
GOCMD = go
GOTEST = $(GOCMD) test -v

test:
	$(GOTEST) ./...

image:
	docker build -t sp0x/go-ddns:latest .

build:
	go build -o $(NAME) -ldflags "-s -w" ./rest-api

console:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --rm sp0x/go-ddns:latest bash

devconsole:
	docker run -it --rm -v ${PWD}/rest-api:/usr/src/app -w /usr/src/app golang:1.8.5 bash

server_test:
	docker run -it -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --rm sp0x/go-ddns:latest

api_test:
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	dig @localhost foo.example.org

api_test_46:
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=1.2.3.4"
	curl "http://localhost:8080/update?secret=changeme&domain=foo&addr=2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	dig @localhost foo.example.org
	dig @localhost AAAA foo.example.org

api_test_multiple_domains:
	curl "http://localhost:8080/update?secret=changeme&domain=foo,bar,baz&addr=1.2.3.4"
	dig @localhost foo.example.org
	dig @localhost bar.example.org
	dig @localhost baz.example.org

api_test_invalid_params:
	curl "http://localhost:8080/update?secret=changeme&addr=1.2.3.4"
	dig @localhost foo.example.org

api_test_recursion:
	dig @localhost google.com

deploy: image
	docker run -it -d -p 8080:8080 -p 53:53 -p 53:53/udp --env-file envfile --name=dyndns sp0x/go-ddns:latest
