COMMIT_SHA?=$(shell git describe --match=xxx --always --abbrev=40 --dirty)

lint-deps: dep-staticcheck

lint: lint-deps
	staticcheck -checks=all -tests ./...

dep-staticcheck:
	@command -v staticcheck >/dev/null 2>&1 || (echo "missing staticcheck"; GO111MODULE=off go get honnef.co/go/tools/cmd/staticcheck)

.PHONY: build
build:
	gcloud builds submit --substitutions=_LOCATION=us-east4,_CLUSTER=apps-cluster-1,COMMIT_SHA=$(COMMIT_SHA)

