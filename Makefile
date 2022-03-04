generate:
	@gqlgen generate
run:
	@go run main.go serve --introspection=true