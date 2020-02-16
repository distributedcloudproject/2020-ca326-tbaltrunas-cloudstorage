# This assumes that we have done the proper linking
# i.e. we need a symlink from local project files to $GOPATH
# cd $GOPATH/src
# ln -s /project/dir/name name


PROJECT_NAME=cloud
PROJECT_BINARY=cloud
LOGS_DIRNAME=logs


deps:
	go get github.com/hashicorp/logutils


make: deps
	go build $(PROJECT_NAME)


clean:
	rm -f $(PROJECT_BINARY)


fclean: clean
	go clean $(PROJECT_NAME)
	rm -f "Save File"
	rm -rf $(LOGS_DIRNAME)
	rm -rf **/$(LOGS_DIRNAME)


re: fclean make


run: re
	# run as: make run ARGS="args"
	./$(PROJECT_BINARY) $(ARGS)


test: make
	# run with ARGS="-v" flag for verbose mode
	@go test ./... $(ARGS)


deploy: make
	mv cloud /bin
