# orphanroutepolicycleaner

Deletes orphan Cloud Foundry route policies.

## What it does

1. Lists all route policies (`/v3/route_policies`) using the go-cfclient v3 library.
2. Collects the `source` of every route policy and deduplicates them. A source looks like:
   ```
   cf:space:8bbdb186-5854-4e6c-a1b4-f6382a324a51
   cf:app:b2b74671-9179-4a1a-95b7-582e9918efd4
   cf:org:856dc2bc-b870-444c-9ce9-a6a0ea432ae5
   ```
3. For each unique source it queries the cloud controller to check if the org / space / app still exists.
4. Every route policy whose source no longer exists is deleted (the delete job is polled until completion).

If the existence check fails for another reason than "not found" (e.g. a network error), or the source
format/type is not recognized, the source is considered to still exist so nothing is deleted.

## Configuration (environment variables)

| Variable | Required | Default | Description                                                     |
|---|---|---|-----------------------------------------------------------------|
| `CF_API_ADDR` | yes | | Cloud Controller API url, e.g. `https://api.sys.your.cf.domain` |
| `CF_USERNAME` | yes | | UAA client id (client credentials)                              |
| `CF_PASSWORD` | yes | | UAA client secret                                               |
| `SKIP_SSL_VALIDATION` | no | `false` | Skip TLS certificate validation                                 |
| `DRY_RUN` | no | `false` | When `true`, only report what would be deleted                  |
| `DEBUG` | no | `false` | When `true`, be more verbose on what is found and done          |
