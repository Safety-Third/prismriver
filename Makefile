all: build

build:
	fileb0x assets.json
	go build cmd/prismriver/prismriver.go

install:
	install -b -D -m644 "prismriver.yml" "/etc/prismriver/prismriver.yml"
	install -D -m755 "prismriver" "/usr/local/bin/prismriver"
	install -D -m644 "prismriver.service" "/usr/lib/systemd/system/prismriver.service"
	install -D -m644 "prismriver-user.service" "/usr/lib/systemd/user/prismriver.service"

run: build
	./prismriver

validate:
	./scripts/validate.sh

.PHONY: all build install run validate
