NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m
DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
ECHO=echo
else
ECHO=/bin/echo -e
endif

all: deps
	@mkdir -p bin/
	@$(ECHO) "$(OK_COLOR)==> Building$(NO_COLOR)"
	@bash --norc -i ./scripts/devcompile.sh

deps:
	@$(ECHO) "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@echo $(DEPS) | xargs -n1 go get -d

updatedeps:
	@$(ECHO) "$(OK_COLOR)==> Updating all dependencies$(NO_COLOR)"
	@go get -d -v -u ./...
	@echo $(DEPS) | xargs -n1 go get -d -u

clean:
	@rm -rf bin/ local/ pkg/ src/ website/.sass-cache website/build

format:
	go fmt ./...

test: deps
	@$(ECHO) "$(OK_COLOR)==> Testing Packer...$(NO_COLOR)"
	go test ./...

.PHONY: all clean deps format test updatedeps

docker-gopath:
	@test -z "`docker ps -a | grep gopath`" && \
	docker run -d --name gopath -v /gopath ubuntu:14.04 true || \
	$(ECHO) "$(OK_COLOR)==> 'gopath' volume container already exists$(NO_COLOR)"

docker-deps: docker-gopath
	@docker run -t -i --rm=true \
	--volumes-from gopath \
	-v `pwd`:/gopath/src/github.com/mitchellh/packer \
	-m=1g \
	google/golang \
	/bin/bash -c 'go get -u github.com/mitchellh/gox && cd /gopath/src/github.com/mitchellh/packer && make deps'

docker: docker-gopath
	@docker run -t -i --rm=true \
	--volumes-from gopath \
	-v `pwd`:/gopath/src/github.com/mitchellh/packer \
	-m=1g \
	google/golang \
	/bin/bash -c 'cd /gopath/src/github.com/mitchellh/packer && make'

docker-test: docker-gopath
	@docker run -t -i --rm=true \
	--volumes-from gopath \
	-v `pwd`:/gopath/src/github.com/mitchellh/packer \
	-m=1g \
	google/golang \
	/bin/bash -c 'cd /gopath/src/github.com/mitchellh/packer && make test'
