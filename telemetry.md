# Telemetry

Telemetry is a unique API endpoint, distinct from the other resources in the API. It is used to submit logs from
a backup attempt. This is a separate endpoint from the backup attempts themselves, as the logs may not be stored
in the same way as the backup attempts. This is to allow for more flexibility in the storage of logs, and to
allow for more detailed logs to be submitted than would be reasonable to store in the database.

## Protocol

The telemetry endpoint is a RESTful API, using the HTTP protocol. It is a POST request, with the body of the
request as the logs themselves. The format of the logs is expected to be a JSON object per log line.

Each submission is expected to be one JSON object per line, with the following fields:

* `run`: The run ID of the backup attempt that this log is for. UUID format.
* `timestamp`: The timestamp of the line. ISO 8601 format.
* `level`: The log level of the line. One of `DEBUG`, `INFO`, `WARNING`, `ERROR`, `CRITICAL`.
* `fields`: A JSON object of additional fields. This is optional, and can be used for additional metadata.
* `message`: The message of the log line.

Multiple lines can be sent in a single submission, but each line must be a separate JSON object.

Submissions are via `POST` to the appropriate endpoint. As of this writing, the endpoint is
`/telemetry/{instance}/logs`, where `{instance}` is the instance ID that the logs are for. However,
the [API spec](./api.yaml) is authoritative, and should be consulted for the correct endpoint.

Each successful submission returns a `201 Created` response.

A particular implementation may have a maximum size for the body of the request. If the body is too large,
the server should return a `413 Request Entity Too Large` response.

