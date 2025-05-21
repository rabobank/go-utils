### serviceinstancecleaner

A simple program that iterates over cf service instances of the given service_offering_name, checks if they are bound to any apps and if not, deletes them, or sends notifications.

## Environment variables:

* `CF_API_ADDR` - The Cloud Foundry API endpoint (https://api.sys.blabla.com). This environment variable is required.
* `CF_USERNAME` - The Cloud Foundry username. This environment variable is required.
* `CF_PASSWORD` - The Cloud Foundry password. This environment variable is required.
* `SERVICE_OFFERING` - The name of the service offering to check for (rds-service, config-hub, credhub. This environment variable is required.
* `RUN_TYPE` - Should be either `stopdaily` , `stopweekly` or `stopdaily,stopweekly`. If an app has it's AUTOSTOP label set to one of those values, it will be processed. This environment variable is required.
* `DRY_RUN` - If set to `true`, the program will only print the service instances that would be eligible, but not actually delete them, defaults to `false`.
* `SKIP_SSL_VALIDATION` - defaults to `false`.
* `EXCLUDED_ORGS` - A comma separated list of orgs to exclude from the process, defaults to `system`.
* `EXCLUDED_SPACES` - A comma separated list of spaces to exclude from the process, defaults to `""`.
* `GRACE_PERIOD` - The number of days to wait (uses the instance's and binding's updated_at field) before stopping/deleting an app. This environment variable is required.

Create the uaa user that runs the serviceinstancecleaner program:
```
uaac target https://uaa.sys.cfd05.rabobank.nl
uaac token client get admin -s <uaa admin client pasword>
uaac client add cf-housekeeping --name cf-housekeeping --authorized_grant_types client_credentials,refresh_token --authorities doppler.firehose,cloud_controller.admin
```
