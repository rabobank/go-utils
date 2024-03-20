### appstopper

A simple program that iterates over cf apps, checks their labels and stops the apps optionally.  

## Environment variables:

* `CF_API_ADDR` - The Cloud Foundry API endpoint (https://api.sys.blabla.com). This environment variable is required.
* `CF_USERNAME` - The Cloud Foundry username. This environment variable is required.
* `CF_PASSWORD` - The Cloud Foundry password. This environment variable is required.
* `RUN_TYPE` - Should be either `daily` , `weekly` or `daily,weekly`. If an app has it's AUTOSTOP label set to one of those values, it will be processed. This environment variable is required.
* `DRY_RUN` - If set to `true`, the program will only print the apps that would be eligible, but not actually scale/restart them, defaults to `false`.
* `SKIP_SSL_VALIDATION` - defaults to `false`.
* `EXCLUDED_ORGS` - A comma separated list of orgs to exclude from the process, defaults to `system`.
* `EXCLUDED_SPACES` - A comma separated list of spaces to exclude from the process, defaults to `""`.

If you want an app to be excluded from the process, you can add the following labels to it:  

* do not set a label or set the label AUTOSTOP=daily - if you want your app to be automatically stopped every day 
* set the label AUTOSTOP=none - if you don't want your app to be stopped at all 
* set the label AUTOSTOP=weekly - if you want your app to be stopped every week

You can set/un-set a label with the cf command:
 ```
cf set-label   app <my-app> AUTOSTOP=daily    # set a label
cf unset-label app <my-app> AUTOSTOP          # remove a label
 ```

Create the uaa user that runs the appstopper program:
```
uaac target https://uaa.sys.cfd05.rabobank.nl
uaac token client get admin -s <uaa admin client pasword>
uaac client add cf-housekeeping --name cf-housekeeping --authorized_grant_types client_credentials,refresh_token --authorities doppler.firehose,cloud_controller.admin
```