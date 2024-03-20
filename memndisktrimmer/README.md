### memndisktrimmer

A simple program that iterates over cf apps, their processes and stats, finding the apps that use much less memory and disk than they allocate, it then rescales and restarts those apps.  

## Environment variables:

* `CF_API_ADDR` - The Cloud Foundry API endpoint (https://api.sys.blabla.com). This environment variable is required.
* `CF_USERNAME` - The Cloud Foundry username. This environment variable is required.
* `CF_PASSWORD` - The Cloud Foundry password. This environment variable is required.
* `MEM_USAGE_THRESHOLD_PERCENTAGE` - If an app uses less than this percentage of the allocated memory, it will be eligible, defaults to 20.
* `DISK_USAGE_THRESHOLD_PERCENTAGE` - If an app uses less than this percentage of the allocated disk, it will be eligible, defaults to 20.
* `MEM_SCRAPE_PERCENTAGE` - The percentage with which the allocated memory is reduced, defaults to 20.
* `DISK_SCRAPE_PERCENTAGE` - The percentage with which the allocated disk is reduced, defaults to 20.
* `JAVA_MINIMUM_MB` - If it's a Java app and has less than this, we will not touch it (java can require much more memory during startup), defaults to 768.
* `LAST_UPDATED_AGE_THRESHOLD` - If an app was updated less than this number of days ago, it is not eligible, defaults to 5. (there are apps that are deployed very often, scaling them down each time is not a good idea)
* `DRY_RUN` - If set to `true`, the program will only print the apps that would be eligible, but not actually scale/restart them, defaults to `false`.
* `SKIP_SSL_VALIDATION` - defaults to `false`.
* `EXCLUDED_ORGS` - A comma separated list of orgs to exclude from the process, defaults to `system`.
* `EXCLUDED_SPACES` - A comma separated list of spaces to exclude from the process, defaults to `""`.

If you want an app to be excluded from the memndisktrimmer, simply add the label `NO_MEMNDISK_TRIM` to the app with the value `true`: ```cf set-label <my-app> NO_MEMORY_TRIM=true```