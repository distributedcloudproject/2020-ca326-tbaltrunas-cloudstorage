PROJECT_NAME=cloud
PROJECT_BINARY=cloud


make:
	@go build $(PROJECT_NAME)


clean:
	@rm -f $(PROJECT_BINARY)


fclean: clean
	@go clean $(PROJECT_NAME)


re: fclean make


run: re
	# run as: make run ARGS="args"
	@./$(PROJECT_BINARY) $(ARGS)


test:
	@go test ./...
