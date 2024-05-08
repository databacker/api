# go api

Contains the API for go, as generated from the openapi3 spec in the root directory of this
repository.

These files should be generated using `make go` from the root directory.

Types will be generated in the `github.com/databacker/api/go/api` package, and should be imported
from there.

The dual "api" in the module name - `api/go/api` - is intentional. `github.com/databacker/api` is
the repository, `go/` is for the generated go files, as others may follow, and `api/` is so that
the package name is `api` and not the generic `go`.
