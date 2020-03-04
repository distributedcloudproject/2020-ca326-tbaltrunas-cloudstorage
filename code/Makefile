# This assumes that we have done the proper linking
# i.e. we need a symlink from local project files to $GOPATH
# cd $GOPATH/src
# ln -s /project/dir/name name


PROJECT_NAME=cloud
PROJECT_BINARY=cloud
LOGS_DIRNAME=logs


make: deps
	go build $(PROJECT_NAME)


deps:
	go get github.com/hashicorp/logutils
	go get github.com/ricochet2200/go-disk-usage/du
	go get github.com/fsnotify/fsnotify
	go get fyne.io/fyne
	go get github.com/sqweek/dialog


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
	# run with ARGS="args" for extra test arguments
	# "-v" for verbose mode
	# "-run TestFunctionName" to only run the test function TestFunctionName
	# add environment variables in form "make test ... KEY1=VALUE1 KEY2=VALUE2 ..."
	@go test ./... $(ARGS)


deploy: make
	mv cloud /bin
