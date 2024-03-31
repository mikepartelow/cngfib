jaeger:
	docker run --name jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4317:4317 -p 4318:4318 jaegertracing/all-in-one:1.55

bench:
	go clean -testcache
	# go test -bench ./... should work, but doesn't
	cd pkg/fib/ && go test -bench .

test:
	go test -race -cover -v ./...

run: test
	go run main.go
