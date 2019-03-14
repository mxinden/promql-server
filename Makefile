SWAGGER = docker run \
	--user=$(shell id -u $(USER)):$(shell id -g $(USER)) \
	--rm \
	-v $(shell pwd):/go/src/github.com/mxinden/promql-server \
	-w /go/src/github.com/mxinden/promql-server quay.io/goswagger/swagger:v0.18.0

api/v1/models api/v1/restapi: api/v1/openapi.yaml
	-rm -r api/v1/{models,restapi}
	$(SWAGGER) generate server -f api/v1/openapi.yaml --exclude-main -A promql-server --target api/v1/
