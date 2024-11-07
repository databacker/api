# api

API specification for communicating between backup services and a management service

databacker software backs up databases to remote targets, e.g. files, CMB shares,
S3 buckets, etc.

This is the specification of the API between a databacker instance and a databacker controller.

You do **not** need to use this API or the controller. Each databacker instance
is configured locally using environment variables, CLI flags or a configuration file either:

* for its entire configuration; OR
* to connect to a controller to receive its configuration

If you do not use an API-compliant controller, this API is not relevant to your use case.

You **may** choose to use a controller to manage: targets and scheduling; for reporting; or both.

## Services

The controller optionally provides up to two services:

* scheduling
* reporting

### Scheduling

If scheduling is enabled, the local databacker instance will not use local configuration, except to
connect with its controller. Instead,
it will connect to the controller to receive all of its information: targets, schedules, etc. The only
locally-provided information is the address of the controller and credentials.

### Reporting

If reporting is enabled, each run of a backup will send reports to the controller, which can provide aggregated backup
information as well as notifications of success, failure, time to run, and other useful metrics.
Secrets, such as credentials, wil *not* be reported, while database names, hostnames and other
information will be reported.

Reporting is broken down into two parts:

* logs
* events

Logs are output logs as provided by the databacker instance, e.g. `stdout` and `stderr` output. These usually
are human-readable.

Events are individual events of the backup, including structured data. They include start time, script execution times,
success or failure, and other structured data. There is some duplication between logs and events, but events
are explicitly intended to indicate state and provide timing metrics, without having to parse text logs.
This also frees up logging to include things that might or might not make sense to a parser, without having to worry
about breaking a parser.

## API Specification

The API is a RESTful API, with JSON payloads.
The API is defined in this repository in [OpenAPI 3.0 spec](https://github.com/OAI/OpenAPI-Specification) format.
It is used to generate bindings both for a controller implementation and databacker instances.

The API specification includes the following sections:

* reporting (by databacker instances)
* orchestration (of databacker instances)
* administration (by end-users)

### Protocol

The protocol is RESTful, with json payloads over HTTP, secured by TLS.
json payloads are used everywhere, except for streaming large amounts of data,
for example logs, both submission by a databacker endpoint and retrieval by a user.

### Authentication & Authorization

API requests, whether from an end-user or a databacker instance, must be authenticated.
[Json Web Tokens (JWT)](https://jwt.io/), defined in [RFC7519](https://tools.ietf.org/html/rfc7519)
are used for validating and authorizing requests.

The JWTs are issued via one of two methods:

* OAuth2 Authorization Code Grant, for end-users (administration)
* OAuth2 Client Flow, for databacker instances (reporting, orchestration)

#### OAuth2 Authorization Code Grant

End users accessing the controller Web UI use the
[OAuth2 Authorization Code Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.1).

When creating a new databacker instance using the admin UI or API, the user must submit an ECDSA public key
to affiliate with the instance. The user must generate the keypair client-side, and submit the public key.

A Web UI _may_, for user convenience, generate a new keypair client-side, and submit the public key.
This keypair must be generated entirely client-side; the API simply accepts a public key to associate with the
instance. Web UIs and similar should generate the key client side, or allow the user to submit a public key.

This API does not specify the authorization server, only that a valid JWT must exist.
A controller may implement an authorization server, but it is not required as part of this
specification.

#### OAuth2 Client Flow

Databacker instances use the [OAuth2 Client Flow](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4).

The databacker instance authenticates using the ECDSA public key affiliated with the instance, using
it to generate a JWT, which is submitted for all future requests. Once the JWT expires, the databacker
instance must re-authenticate.

### Endpoints

This API spec does not delineate URL endpoints, or distinguish between URL endpoints for various
purposes. The entire spec can be implemented in a single endpoint, or split among multiple.

Databacker instances only need to be configured for a single endpoint providing orchestration.
When connecting to the endpoint providing orchestration, the databacker instance receives
all other connection URLs, including for reporting services, whether distinct from the endpoint
providing orchestration services or the same.

Even if the account is not configured for orchestration, it the orchestration endpoint,
and is provided the reporting endpoint.

Routes in the API are global; a single API scheme is defined for the entire API, independent of
endpoints. Separate endpoints may choose to implement different subsets of the API.

An endpoint that implements only a subset of the API, upon being queried for a route
that is not implemented may return one of the following:

* `301 Moved Permanently` with a `Location` header to the correct endpoint, if it knows of the appropriate location.
* `410 Gone` if it does not know of the appropriate location. This indicates that it is a valid part of the API, but this server does not implement it.
* `404 Not Found` if this is not a known valid part of the API.

### Versioning

The API _as a whole_ is not versioned, e.g. `https://endpoint/v1/` and `https://endpoint/v2/`.
While this may not be possible permanently, this specification shall attempt to avoid it as long
as possible.

As this API is as closely REST compatible as possible, all resources are permanent endpoints,
e.g. `/config/{instance}` and `/report/{instance}`. New resources will be released at new endpoints.

Specific versions of individual resources are versioned via HTTP headers.
The `Accept` header is used to specify the version of the resource requested.
Each endpoint that represents a resource has one or more specific media-type which it returns.

Generic media-types, such as `application/json` are not used. Instead, each resource has a specific
media-type, e.g. `application/vnd.databack.device-config.v1+json`.

The version is specified in the media-type, and the format is specified in the extension, e.g. `+json`.

* If the format is not specified, the default is `json`.
* If the version is not specified, the default is the highest one available.
* If no media-type is provided, the default is the highest version of the default format.

New endpoints or new versions of individual resources require new media types and knowledge of the
new paths, and therefore require new databacker versions. However:

* all new endpoints should be backwards compatible with previous versions of databacker, which will be unaware of them
* some new versions of existing resources should be backwards compatible with previous versions of databacker, when they simply add new fields, that should be ignored by older versions of databacker
* other new versions of existing resources may not be backwards compatible with previous versions of databacker, when they change the meaning of existing fields, or remove fields; new media-types should be used for these
