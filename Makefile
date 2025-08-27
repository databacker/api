all: sdk

.PHONY: all sdk sdk-container go

sdk: go
	@echo Done building SDKs

# sdk-container is a helper target to build the SDKs in a container. It builds a container image
# based on the contents of .devcontainer/Dockerfile and runs the make sdk target inside the container.
sdk-container:
	docker build -f .devcontainer/Dockerfile -t databack-api-builder .
	docker run --rm -v $(PWD):/src -w /src -u $$(id -u) databack-api-builder make sdk

# MUST use oapi-codegen v2.5.0 or later.
go:
	oapi-codegen -config ./cfg-schemas.yml ./src/schemas.yaml
	oapi-codegen -config ./cfg-api.yml ./src/api.yaml
