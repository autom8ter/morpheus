generate:
	@gqlgen generate
run:
	@go run main.go serve

dataloader:
	@cd pkg/dataloader; go run github.com/vektah/dataloaden NodeLoader string *github.com/autom8ter/morpheus/pkg/api.Node

bench-storage:
	@go test -bench=Benchmark ./pkg/storage -benchmem -run=^$

test-storage:
	@go test -v ./pkg/storage

query-size:
	@go run main.go query -f examples/queries/size.graphql

query-login:
	@go run main.go query -f examples/queries/login.graphql