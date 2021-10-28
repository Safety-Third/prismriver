all: build validate

build:
	fileb0x assets.json
	go build cmd/prismriver/prismriver.go

install:
	install -b -D -m644 "config/prismriver.yml" "/etc/prismriver/prismriver.yml"
	install -D -m755 "prismriver" "/usr/local/bin/prismriver"
	install -D -m644 "init/prismriver.service" "/usr/lib/systemd/system/prismriver.service"
	install -D -m644 "init/prismriver-user.service" "/usr/lib/systemd/user/prismriver.service"

run: build
	./prismriver

validate:
	go vet ./...
	staticcheck ./...

.PHONY: all build install run validate
