build:
	go build -tags netgo ./...

release: cleanall
	./build.sh

clean:
	rm -f godeb

cleanall: clean
	rm -f godeb-*.tar.gz godeb-*.tar.gz.asc

.PHONY: build release clean cleanall
