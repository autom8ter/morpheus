generate:
	@gqlgen generate
run:
	@go run main.go serve

dataloader:
	@cd pkg/dataloader; go run github.com/vektah/dataloaden NodeLoader string *github.com/autom8ter/morpheus/pkg/api.Node
