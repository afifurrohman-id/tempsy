MAIN_FILE:=cmd/files/main.go
EXE_NAME:=tempsy
ifeq ($(go env GOOS), windows)
		EXE_NAME=tempsy.exe
endif

build: ${MAIN_FILE}
	CGO_ENABLED=0 go build -o ${EXE_NAME} -ldflags "-w -s" ${MAIN_FILE}

run: ${MAIN_FILE}
	CGO_ENABLED=0 go run -ldflags "-w -s" ${MAIN_FILE}

test: ${MAIN_FILE}
	CGO_ENABLED=1 go test --cover -race -v -ldflags "-w -s" ./...

clean: deployments/compose.yaml
	if [ -f ${EXE_NAME} ]; then rm ${EXE_NAME}; fi

	docker compose -f deployments/compose.yaml down

build-image: build/package/Containerfile
	docker build -f build/package/Containerfile -t tempsy .

git-mod-update:
	git submodule update --remote --merge

fmt:
	go fix ./... && go fmt ./... && go vet ./...

lint-image: build/package/Containerfile
	docker run --rm -i hadolint/hadolint:latest-alpine < build/package/Containerfile
	docker run --rm -i hadolint/hadolint:latest-alpine < build/package/test.Containerfile

compose-up: deployments/compose.yaml
	docker compose -f deployments/compose.yaml up -d

