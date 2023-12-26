run: main.go
	CGO_ENABLED=0 go run -ldflags "-w -s" main.go

EXE_NAME:="tempsy.exe"
ifeq ($(go env GOOS), "windows")
		EXE_NAME="tempsy.exe"
else
		EXE_NAME="tempsy"
endif

build: main.go
	CGO_ENABLED=0 go build -o ${EXE_NAME} -ldflags "-w -s" main.go

test: main.go
	go test -v --cover -race ./...

clean: deployments/compose.yaml
	if [ -f ${EXE_NAME} ]; then rm ${EXE_NAME}; fi

	docker compose -f	deployments/compose.yaml down

build-image: build/Containerfile
	docker build -f build/Containerfile -t tempsy .

compose-up: deployments/compose.yaml
	docker compose -f deployments/compose.yaml up -d

