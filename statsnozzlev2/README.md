!! statsnozzlev2 - V2 Firehose (using RLP Gateway)

The V2 firehose is using the RLP (reverse log proxy) gateway, see [this picture](https://docs.cloudfoundry.org/loggregator/images/architecture/observability.png), it works fine but takes almost triple the amount of CPU resources than the V1 Firehose.  
See also these Pivotal articles:

* https://community.pivotal.io/s/article/Logs-metrics-read-from-Loggregator-V2-Firehose-log-stream-endpoint-drop-under-load
* https://community.pivotal.io/s/article/Consuming-logs-metrics-via-the-RLP-Gateway-V2-Firehose-API-will-show-reduced-performance

The advantage of the V2 firehose that the payload you get contains more data, most importantly the RTR logging contains the org, space and app name, which is very handy for doing analysis.  

The following environment variables are available:
* API_ADDR - the address of the CF API endpoint (https://api.sys.mydomain.com)
* SHARD_ID - the shard ID to use, can be any string as long as it doesn't match another shard ID
* CLIENT_ID - the UAA client ID
* CLIENT_SECRET - the UAA client secret
* PRINT_LOGS - whether to print logs to stdout or not
* STORE_IN_DB - whether to store data in an sqlite database or not (true/false), default is true
* DB_FILE - the path/name of the sqlite database file (e.g. /tmp/statsnozzle.db)
* URI_FILTER - a regular expression to filter URIs, e.g. ^/api/v1/.*
* MAX_MESSAGES - the maximum number of messages to process from the firehose, once these have been processed, the nozzle will stop, and if these are also inserted in the database, the statsnozzle will exit

When started it will suck the firehose, filter out "source_type"="RTR" and optionally log to stdout and/or store data in an sqlite database, see resources/sql/create-table.sql for all the rows that are stored.  
If you read a busy firehose, the statsnozzle will consume a lot of CPU, e.g. a foundation that emits 6000 messages per second will consume 200% CPU on a regular Intel Xeon server, once all the messages (MAX_MESSAGES) have been processed, the nozzle will stop, but it will take some time before all the rows are inserted, during that phase it will drop to around 100% CPU.  
With the URI_FILTER you can optionally filter out more.

This is the output of a run with 420.000 MAX_MESSAGES: 

````
2024/05/17 07:50:45 database file:/tmp/StatsNozzle.db does not exist, creating it...  
2024/05/17 07:51:18 processed 175511 envelopes, inserted 42000 rows, bufferedEnvelopes: 133511  
2024/05/17 07:51:48 processed 350175 envelopes, inserted 84000 rows, bufferedEnvelopes: 266175
2024/05/17 07:51:59 max envelopes reached: 420000, inserted 98142 rows
2024/05/17 07:52:16 processed 420000 envelopes, inserted 126000 rows, bufferedEnvelopes: 294000
2024/05/17 07:52:39 processed 420000 envelopes, inserted 168000 rows, bufferedEnvelopes: 252000
2024/05/17 07:53:02 processed 420000 envelopes, inserted 210000 rows, bufferedEnvelopes: 210000
2024/05/17 07:53:25 processed 420000 envelopes, inserted 252000 rows, bufferedEnvelopes: 168000
2024/05/17 07:53:48 processed 420000 envelopes, inserted 294000 rows, bufferedEnvelopes: 126000
2024/05/17 07:54:10 processed 420000 envelopes, inserted 336000 rows, bufferedEnvelopes: 84000
2024/05/17 07:54:33 processed 420000 envelopes, inserted 378000 rows, bufferedEnvelopes: 42000
2024/05/17 07:54:56 processed 420000 envelopes, inserted 420000 rows, bufferedEnvelopes: 0
2024/05/17 07:54:56 all rows (420000) inserted, exiting...
````

Once you have the data in your database, you can do all kinds of analysis:
* how many requests per second
* what are the slowest requests
* which apps take most requests
* distribution of http methods
* distribution of response codes
And way more...
