### buildpackusage

A simple program that iterates over Cloud Foundry apps/droplets, prints buildpack usage counts, and stores every discovered droplet/buildpack row in a SQLite database.

## Environment variables:

* `CF_API_ADDR` - The Cloud Foundry API endpoint (https://api.sys.blabla.com). This environment variable is required.
* `CF_USERNAME` - The Cloud Foundry username. This environment variable is required.
* `CF_PASSWORD` - The Cloud Foundry password. This environment variable is required.
* `STORE_IN_DB` - Whether to write discovered rows to SQLite. Defaults to `true`.
* `DB_FILE` - Path to the SQLite database file. Defaults to `buildpackusage.db` in the current directory.

## Output

The tool prints the total number of discovered buildpack rows and grouped counts by buildpack name, buildpack name/stack, and buildpack name/version/stack before writing the same rows to SQLite.
