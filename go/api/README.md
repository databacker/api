# go api

Contains the API for go, as generated from the openapi3 spec in the root directory of this
repository.

These files should be generated using `make go` from the root directory.

Types will be generated in the `github.com/databacker/api/go/api` package, and should be imported
from there.

The dual "api" in the module name - `api/go/api` - is intentional. `github.com/databacker/api` is
the repository, `go/` is for the generated go files, as others may follow, and `api/` is so that
the package name is `api` and not the generic `go`.

This requires [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
commit d3a2029448254ffee6dcc0284dbd4aeb2e1cab60 or later,
which is to be included in v2.5.0, as it uses
the support for the config file with `output-options.name-normalizer`. As of this writing,
v2.5.0 is not yet released, so you will need to install it from the `main` branch.

Once v2.5.0 is released, we should switch to semver.

Install via:

```sh
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0
# or, if v2.5.0 is not yet release
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@d3a2029448254ffee6dcc0284dbd4aeb2e1cab60
```
