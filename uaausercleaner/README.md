### uaausercleaner

A simple program that will show uaa users that have a createdTime and lastLogonTime of too long ago. For now this is only an example of how to use the uaa api, and it will only list the users .  

## Environment variables:

* `UAA_API_ADDR` - The Cloud Foundry UAA endpoint (https://uaa.sys.blabla.com). This environment variable is required.
* `UAA_CLIENTID` - The uaa client. This environment variable is required.
* `UAA_CLIENTSECRET` - The uaa client secret. This environment variable is required.
* `SKIP_SSL_VALIDATION` - defaults to `false`.
* `CREATED_DAYS_AGO` - Number of days ago the user was created before it will pass the filter, defaults to 400.
* `LASTLOGON_DAYS_AGO` - Number of days ago the user was last logged in before it will pass the filter, defaults to 400.
* `DELETE_USERS` - If set to `true`, the users that match the filter will be deleted. Defaults to `false`.
