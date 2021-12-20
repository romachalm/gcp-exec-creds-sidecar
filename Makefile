repository := rchalumeau/gec

package := gec
cmd := ./cmd/gcp-exec-creds-sidecar
api := api/${package}.yaml

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor
	GO111MODULE=on go mod tidy

.PHONY: lint
lint:
	golangci-lint version
	GL_DEBUG=linters_output GO111MODULE=on golangci-lint run

.PHONY: generate
generate:
	# install the generator with go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	mkdir -p pkg/${package}
	oapi-codegen \
		--package=${package}  \
		--generate=types,chi-server,spec \
		-o pkg/${package}/${package}.gen.go \
		${api}

.PHONY: local
local:
	go run ${cmd}/main.go -verbose

.PHONY: doc
doc:
	openapi-generator generate -i ${api}-g markdown --skip-validate-spec -o docs

.PHONY: test
test:
	go test -cover ./... -v

.PHONY: build
build:
	KO_DOCKER_REPO=${repository} ko publish ${cmd} --bare

.PHONY: build
push:
	KO_DOCKER_REPO=${repository} ko publish ${cmd} --bare --push

kustomize:
	kustomize build deployments