.PHONY: game-backend

lint:
	golangci-lint run --fix --out-format=line-number --issues-exit-code=0 --config .golangci.yml --color always ./...

