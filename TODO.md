openapi: generate client and server code from api.yaml

What does the API need to include? Condensed list here, to be translated to api.yaml after.

resources: instance, config, log

Actions we need to do (requestor):

- create a new instance (admin)
- create a device config (admin)
- get a list of instances (admin)
- get info about a specific device, other than config (admin) 
- get a device config (admin, device)
- create a backup attempt (device)
- submit logs for a backup attempt (device)
- get information about a backup attempt, e.g. date, device, start, finish, etc., except for actual logs (admin)
- get logs for a backup attempt (admin)

Paths:

`GET /instances/{instance}` - what does this get? I guess the config, but should that not be its own thing? Or maybe we link to `GET /configs/{config}`? Or maybe we just get the instance info, and then `GET /configs/{config}` to get the config? Or alternatively `GET /instances/{instance}/config`? Depends if we should be tracking histories of configs. Also if we need the actual path to handle security, e.g. `GET /configs/{config}` requires determining which instance the config is for, and then if the requestor has access to that config. If we just do `GET /instances/{instance}/config` then we can just check if the requestor has access to the instance; instance _already_ is in the path.

`POST /instances/{instance}/config` or `POST /configs` - create a new config for a specific instance.
This is less of an issue to use either or, as we will process the body to determine which instance it is for. Or conversely, we might make it required in the query. Or simpler may be to include in the path.

`POST /telemetry/{instance}/logs` - submit logs from an individual run for a backup attempt. An alternative is to create a backup attempt, and then submit logs: `POST /backups` (or `POST /instances/{instance}/backups`) to create a backup attempt, and then `POST /backups/{backup}/logs` (or `POST /instances/{instance}/backups/{backup}/logs`) to submit logs.