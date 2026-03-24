### serviceusage

A simple program that iterates over cf service instances of the given service_offering_name(s), and list it's created/updated times, org, space and bound app names 

## Environment variables:

* `CF_API_ADDR` - The Cloud Foundry API endpoint (https://api.sys.blabla.com). This environment variable is required.
* `CF_USERNAME` - The Cloud Foundry username. This environment variable is required.
* `CF_PASSWORD` - The Cloud Foundry password. This environment variable is required.
* `SERVICE_OFFERING` - The name of the service offering to check for (rds-service, config-hub, credhub. This environment variable is required.
* `SKIP_SSL_VALIDATION` - defaults to `false`.

Create the uaa user that runs the serviceusage program:
```
uaac target https://uaa.sys.cfd05.rabobank.nl
uaac token client get admin -s <uaa admin client pasword>
uaac client add cf-housekeeping --name cf-housekeeping --authorized_grant_types client_credentials,refresh_token --authorities doppler.firehose,cloud_controller.admin
```
