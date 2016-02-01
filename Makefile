PACKAGES=$(shell find . -name '*.go' -print0 | xargs -0 -n1 dirname | sort --unique)

deadcode:
	@deadcode $(PACKAGES) 2>&1

tools:
	go get github.com/remyoudompheng/go-misc/deadcode
	go get github.com/alecthomas/gocyclo
	go get github.com/opennota/check/...
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck


cyclo:
	@gocyclo -over 10 $(PACKAGES)

aligncheck:
	@aligncheck $(PACKAGES)

defercheck:
	@defercheck $(PACKAGES)


structcheck:
	@structcheck $(PACKAGES)
test: clean
		go test -v ./... -cover

clean:
		find . -name flymake_* -delete

update:
		rm -rf Godeps/
		find . -iregex .*go | xargs sed -i 's:".*Godeps/_workspace/src/:":g'
		godep save -r ./...

cover-package: clean
		go test -v ./$(p)  -coverprofile=/tmp/coverage.out
		go tool cover -html=/tmp/coverage.out

sloccount:
		 find . -path ./Godeps -prune -o -name "*.go" -print0 | xargs -0 wc -l

install: clean
		go install github.com/xpandmmi/authentication-api
		cd xctl && $(MAKE) install && cd ..

run: install
		authentication-api -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=/authentication -statsdAddr=localhost:8125 -statsdPrefix=authentication -logSeverity=INFO

run-fast: install
		authentication-api -etcd=${ETCD_NODE1} -etcd=${ETCD_NODE2} -etcd=${ETCD_NODE3} -etcdKey=/authentication

docker-clean:
		docker rm -f authentication-api

build:
		./build.py
