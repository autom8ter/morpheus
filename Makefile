generate:
	@gqlgen generate
run:
	@go run main.go serve

dataloader:
	@cd pkg/dataloader; go run github.com/vektah/dataloaden NodeLoader string *github.com/autom8ter/morpheus/pkg/api.Node

bench-persist:
	@go test -bench=Benchmark ./pkg/persistence -benchmem -run=^$

test-persist:
	@go test -v ./pkg/persistence

query-size:
	@go run main.go query -f examples/queries/size.graphql

query-login:
	@go run main.go query -f examples/queries/login.graphql