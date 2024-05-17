### StatsNozzle

Listens on the firehose (V1) for events and counts them, by EventType, Origin, Job, Deployment and IP. Every 5 seconds it outputs the numbers to stdout.  
You can optionally store the stats in an SQLite database.

Environment variables used:

* API_ADDR - the address of the Cloud Foundry API (https://api.sys.mydomain.com)
* CF_USERNAME - the username to use for the Cloud Foundry API
* CF_PASSWORD - the password to use for the Cloud Foundry API
* STORE_IN_DB - if set to true, the stats will be stored in an SQLite database
* PRINT_LOGS - if set to true, the logs will be printed to stderr
* PRINT_STATS - if set to true, the stats will be printed to stdout every STATS_INTERVAL seconds
* STATS_INTERVAL - the interval in seconds at which the stats will be printed to stdout
* STATS_SIZE - the number of entries in the stats output (top xx entries)
* URI_FILTER - a regular expression to filter the URIs to count
