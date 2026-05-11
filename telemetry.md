# Telemetry

Telemetry is a unique API endpoint, distinct from the other resources in the API. It is used to submit logs and
OpenTelemetry traces from a backup attempt. These are separate endpoints from the backup attempts themselves,
as the telemetry data may not be stored in the same way as backup records. This allows for more flexibility in
storage and richer detail than would be reasonable to store in the database.

---

## Logs

### Protocol

The log endpoint is a RESTful API, using HTTP. Submissions are `POST` requests with the body containing
one JSON object per line.

Each line must be a JSON object with the following fields:

* `run`: The run ID of the backup attempt that this log is for. UUID format.
* `timestamp`: The timestamp of the line. ISO 8601 format.
* `level`: The log level of the line. One of `DEBUG`, `INFO`, `WARNING`, `ERROR`, `CRITICAL`.
* `fields`: A JSON object of additional fields. This is optional, and can be used for additional metadata.
* `message`: The message of the log line.

Multiple lines can be sent in a single submission, but each line must be a separate JSON object.

Submissions are via `POST` to the appropriate endpoint. As of this writing, the endpoint is
`/telemetry/{instance}/log`, where `{instance}` is the instance ID that the logs are for. However,
the [API spec](./api.yaml) is authoritative, and should be consulted for the correct endpoint.

Each successful submission returns a `201 Created` response.

A particular implementation may have a maximum size for the body of the request. If the body is too large,
the server should return a `413 Request Entity Too Large` response.

---

## Traces — Backup Span Semantic Convention

Backup engines send traces via the `/telemetry/{instance}/traces` endpoint using
[OTLP](https://opentelemetry.io/docs/specs/otlp/) (serialized `ExportTraceServiceRequest` protobuf).

This section defines the **stable semantic convention** that all backup engines MUST follow so that the
cloud service can reliably derive `BackupTimelineEvent` records from stored spans.

The Go types for all names and keys below are generated from [`schemas.yaml`](./src/schemas.yaml) into
`github.com/databacker/api/go/api` and can be imported directly.

---

### Span Names

Each phase of a backup run is represented by a single OTEL span. The span name MUST be one of the
`BackupSpanName` enum values:

| Span name  | Go constant          | Description                                         |
|------------|----------------------|-----------------------------------------------------|
| `run`      | `BackupSpanRun`      | Root span covering the full backup run              |
| `connect`  | `BackupSpanConnect`  | Establishing the database connection                |
| `snapshot` | `BackupSpanSnapshot` | Creating a consistent snapshot / transaction        |
| `dump`     | `BackupSpanDump`     | Serialising / dumping database data                 |
| `upload`   | `BackupSpanUpload`   | Uploading the backup artifact to the target         |
| `verify`   | `BackupSpanVerify`   | Verifying the uploaded artifact                     |
| `cleanup`  | `BackupSpanCleanup`  | Removing temporary files and releasing resources    |

All phase spans MUST be children of the root `run` span. The `run` span starts when the backup begins
and ends only after all child spans have completed.

---

### Span Attributes

The following attributes MUST or SHOULD be set on spans. Attribute key constants are defined as
`BackupAttributeKey` enum values in the generated Go package.

#### Attributes present on ALL spans

| Attribute key   | Go constant         | Type   | Required | Description                                           |
|-----------------|---------------------|--------|----------|-------------------------------------------------------|
| `backup.run_id` | `BackupAttrRunID`   | string | MUST     | Stable UUID identifying the backup run                |
| `backup.phase`  | `BackupAttrPhase`   | string | MUST     | Phase name (one of `BackupPhase` enum values)         |
| `backup.status` | `BackupAttrStatus`  | string | MUST     | `running` at span start; `ok` or `error` at span end  |
| `otel.status_code` | `BackupAttrOtelStatusCode` | string | MUST | `OK`, `ERROR`, or `UNSET` per OTEL spec        |
| `otel.status_description` | `BackupAttrOtelStatusDescription` | string | SHOULD on error | Human-readable error description |
| `backup.event.label`   | `BackupAttrEventLabel`   | string | SHOULD | Short user-facing label for the phase          |
| `backup.event.message` | `BackupAttrEventMessage` | string | SHOULD | User-facing status or error message            |

#### Attributes on database-related spans (`connect`, `snapshot`, `dump`)

| Attribute key        | Go constant                  | Type    | Required | Description                              |
|----------------------|------------------------------|---------|----------|------------------------------------------|
| `db.system`          | `BackupAttrDBSystem`         | string  | MUST     | DB system per OTEL semconv (e.g. `postgresql`) |
| `db.name`            | `BackupAttrDBName`           | string  | MUST     | Name of the database being backed up     |
| `server.address`     | `BackupAttrServerAddress`    | string  | MUST     | Hostname or IP of the database server    |
| `server.port`        | `BackupAttrServerPort`       | integer | MUST     | Port number of the database server       |
| `network.transport`  | `BackupAttrNetworkTransport` | string  | SHOULD   | Transport protocol (e.g. `tcp`)          |

#### Attributes on data-volume spans (`dump`, `upload`, `verify`)

| Attribute key           | Go constant              | Type    | Required | Description                                       |
|-------------------------|--------------------------|---------|----------|---------------------------------------------------|
| `backup.bytes`          | `BackupAttrBytes`        | integer | SHOULD   | Bytes processed in this phase                     |
| `backup.object_count`   | `BackupAttrObjectCount`  | integer | SHOULD   | Objects (tables, files, etc.) processed           |

#### Target identity attributes (`upload` spans)

These attributes identify **which target** received the backup artifact. They MUST be set on every
`upload` span. When a backup run has multiple targets, emit **one child `upload` span per target**,
each carrying its own set of `backup.target.*` attributes (see [Multiple targets](#multiple-targets)).

Credentials, access keys, tokens, signed URLs, and any other secrets MUST NOT appear in
`backup.target.url` or any other trace attribute.

| Attribute key          | Go constant              | Type   | Required | Description                                                                                              |
|------------------------|--------------------------|--------|----------|----------------------------------------------------------------------------------------------------------|
| `backup.target.name`   | `BackupAttrTargetName`   | string | MUST     | Target key/name from the engine config, e.g. `local-dev`, `daily-s3`, `archive`                        |
| `backup.target.type`   | `BackupAttrTargetType`   | string | MUST     | Target type matching the engine config: `file`, `s3`, or `smb`                                          |
| `backup.target.url`    | `BackupAttrTargetURL`    | string | MUST     | Safe display URL identifying the destination, e.g. `file:///var/lib/databacker/backups`, `s3://bucket/path`, `smb://server/share/path`. No secrets. |

#### Attributes on the root `run` span only

| Attribute key        | Go constant             | Type    | Required | Description                                        |
|----------------------|-------------------------|---------|----------|----------------------------------------------------|
| `backup.exit_code`   | `BackupAttrExitCode`    | integer | SHOULD   | Engine process or result code at run completion    |

---

### BackupPhase vs Span Name

`backup.phase` MUST always match the span name for phase spans. On the root `run` span, `backup.phase`
transitions from `run` (while running) to `complete` once all child phases finish successfully.

---

### Multiple targets

When a backup run uploads to more than one target, engines MUST create **one child `upload` span per
target** rather than a single span that aggregates all targets. Each child span:

- MUST have span name `upload` (`BackupSpanUpload`).
- MUST set `backup.phase` = `upload` (`BackupPhaseUpload`).
- MUST set its own `backup.target.name`, `backup.target.type`, and `backup.target.url`.
- MAY set `backup.bytes` independently per target.
- MUST be a child of the root `run` span (or an intermediate `upload` grouping span).

This allows the cloud service to surface per-target status in `BackupTimelineEvent` records.

---

### Examples

#### Successful backup — root `run` span

```json
{
  "name": "run",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "00f067aa0ba902b7",
  "startTimeUnixNano": "1715000000000000000",
  "endTimeUnixNano":   "1715000180000000000",
  "status": { "code": "STATUS_CODE_OK" },
  "attributes": [
    { "key": "backup.run_id",           "value": { "stringValue": "550e8400-e29b-41d4-a716-446655440000" } },
    { "key": "backup.phase",            "value": { "stringValue": "complete" } },
    { "key": "backup.status",           "value": { "stringValue": "ok" } },
    { "key": "backup.exit_code",        "value": { "intValue": "0" } },
    { "key": "backup.event.label",      "value": { "stringValue": "Backup complete" } },
    { "key": "otel.status_code",        "value": { "stringValue": "OK" } }
  ]
}
```

#### Successful backup — `dump` child span

```json
{
  "name": "dump",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "a3ce929d0e0e4736",
  "parentSpanId": "00f067aa0ba902b7",
  "startTimeUnixNano": "1715000010000000000",
  "endTimeUnixNano":   "1715000120000000000",
  "status": { "code": "STATUS_CODE_OK" },
  "attributes": [
    { "key": "backup.run_id",         "value": { "stringValue": "550e8400-e29b-41d4-a716-446655440000" } },
    { "key": "backup.phase",          "value": { "stringValue": "dump" } },
    { "key": "backup.status",         "value": { "stringValue": "ok" } },
    { "key": "db.system",             "value": { "stringValue": "postgresql" } },
    { "key": "db.name",               "value": { "stringValue": "myapp" } },
    { "key": "server.address",        "value": { "stringValue": "db.internal" } },
    { "key": "server.port",           "value": { "intValue": "5432" } },
    { "key": "network.transport",     "value": { "stringValue": "tcp" } },
    { "key": "backup.bytes",          "value": { "intValue": "104857600" } },
    { "key": "backup.object_count",   "value": { "intValue": "42" } },
    { "key": "backup.event.label",    "value": { "stringValue": "Dump complete" } },
    { "key": "otel.status_code",      "value": { "stringValue": "OK" } }
  ]
}
```

#### Successful backup — `upload` span to a file target

```json
{
  "name": "upload",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "c4de929d0e1f5831",
  "parentSpanId": "00f067aa0ba902b7",
  "startTimeUnixNano": "1715000120000000000",
  "endTimeUnixNano":   "1715000135000000000",
  "status": { "code": "STATUS_CODE_OK" },
  "attributes": [
    { "key": "backup.run_id",        "value": { "stringValue": "550e8400-e29b-41d4-a716-446655440000" } },
    { "key": "backup.phase",         "value": { "stringValue": "upload" } },
    { "key": "backup.status",        "value": { "stringValue": "ok" } },
    { "key": "backup.target.name",   "value": { "stringValue": "local-dev" } },
    { "key": "backup.target.type",   "value": { "stringValue": "file" } },
    { "key": "backup.target.url",    "value": { "stringValue": "file:///var/lib/databacker/dev/database1" } },
    { "key": "backup.bytes",         "value": { "intValue": "104857600" } },
    { "key": "backup.event.label",   "value": { "stringValue": "Upload complete" } },
    { "key": "otel.status_code",     "value": { "stringValue": "OK" } }
  ]
}
```

#### Successful backup — `upload` span to an S3 target

```json
{
  "name": "upload",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "d5ef030e1f2g6942",
  "parentSpanId": "00f067aa0ba902b7",
  "startTimeUnixNano": "1715000120000000000",
  "endTimeUnixNano":   "1715000150000000000",
  "status": { "code": "STATUS_CODE_OK" },
  "attributes": [
    { "key": "backup.run_id",        "value": { "stringValue": "550e8400-e29b-41d4-a716-446655440000" } },
    { "key": "backup.phase",         "value": { "stringValue": "upload" } },
    { "key": "backup.status",        "value": { "stringValue": "ok" } },
    { "key": "backup.target.name",   "value": { "stringValue": "daily-s3" } },
    { "key": "backup.target.type",   "value": { "stringValue": "s3" } },
    { "key": "backup.target.url",    "value": { "stringValue": "s3://my-backup-bucket/prod/database1" } },
    { "key": "backup.bytes",         "value": { "intValue": "104857600" } },
    { "key": "backup.event.label",   "value": { "stringValue": "Upload complete" } },
    { "key": "otel.status_code",     "value": { "stringValue": "OK" } }
  ]
}
```

#### Failed backup — `connect` span (connection refused)

```json
{
  "name": "connect",
  "traceId": "5a3f1c2b8e7d4f6a9c0b1d2e3f4a5b6c",
  "spanId": "b1c2d3e4f5a6b7c8",
  "parentSpanId": "a1b2c3d4e5f6a7b8",
  "startTimeUnixNano": "1715000000000000000",
  "endTimeUnixNano":   "1715000005000000000",
  "status": {
    "code": "STATUS_CODE_ERROR",
    "message": "connection refused: db.internal:5432"
  },
  "attributes": [
    { "key": "backup.run_id",              "value": { "stringValue": "661f9511-f30c-52e5-b827-557766551111" } },
    { "key": "backup.phase",               "value": { "stringValue": "connect" } },
    { "key": "backup.status",              "value": { "stringValue": "error" } },
    { "key": "db.system",                  "value": { "stringValue": "postgresql" } },
    { "key": "db.name",                    "value": { "stringValue": "myapp" } },
    { "key": "server.address",             "value": { "stringValue": "db.internal" } },
    { "key": "server.port",                "value": { "intValue": "5432" } },
    { "key": "network.transport",          "value": { "stringValue": "tcp" } },
    { "key": "backup.event.label",         "value": { "stringValue": "Connection failed" } },
    { "key": "backup.event.message",       "value": { "stringValue": "connection refused: db.internal:5432" } },
    { "key": "otel.status_code",           "value": { "stringValue": "ERROR" } },
    { "key": "otel.status_description",    "value": { "stringValue": "connection refused: db.internal:5432" } }
  ]
}
```

---

### Using the Go constants

```go
import (
    "go.opentelemetry.io/otel/attribute"
    api "github.com/databacker/api/go/api"
)

// Start a child span for the dump phase
ctx, span := tracer.Start(ctx, string(api.BackupSpanDump))
span.SetAttributes(
    attribute.String(string(api.BackupAttrRunID),    runID),
    attribute.String(string(api.BackupAttrPhase),    string(api.BackupPhaseDump)),
    attribute.String(string(api.BackupAttrStatus),   string(api.BackupStatusRunning)),
    attribute.String(string(api.BackupAttrDBSystem), "postgresql"),
    attribute.String(string(api.BackupAttrDBName),   dbName),
    attribute.String(string(api.BackupAttrServerAddress), dbHost),
    attribute.Int(string(api.BackupAttrServerPort),  dbPort),
)
// ... do work ...
span.SetAttributes(
    attribute.String(string(api.BackupAttrStatus),         string(api.BackupStatusOK)),
    attribute.String(string(api.BackupAttrOtelStatusCode), string(api.OtelStatusCodeOK)),
    attribute.Int64(string(api.BackupAttrBytes),           bytesWritten),
    attribute.Int(string(api.BackupAttrObjectCount),       tableCount),
)
span.End()

// One upload span per target — no secrets in the URL
for _, t := range targets {
    _, uploadSpan := tracer.Start(ctx, string(api.BackupSpanUpload))
    uploadSpan.SetAttributes(
        attribute.String(string(api.BackupAttrRunID),       runID),
        attribute.String(string(api.BackupAttrPhase),       string(api.BackupPhaseUpload)),
        attribute.String(string(api.BackupAttrStatus),      string(api.BackupStatusRunning)),
        attribute.String(string(api.BackupAttrTargetName),  t.Name),
        attribute.String(string(api.BackupAttrTargetType),  t.Type),
        attribute.String(string(api.BackupAttrTargetURL),   t.DisplayURL), // MUST NOT contain credentials
    )
    // ... upload ...
    uploadSpan.SetAttributes(
        attribute.String(string(api.BackupAttrStatus),         string(api.BackupStatusOK)),
        attribute.String(string(api.BackupAttrOtelStatusCode), string(api.OtelStatusCodeOK)),
        attribute.Int64(string(api.BackupAttrBytes),           bytesUploaded),
    )
    uploadSpan.End()
}
```

