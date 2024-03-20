### Hazelcast map cleaner

A simple utility that will delete all Hazelcast distributed maps with 0 entries.

## Environment variables:

* `HZ_CLUSTERADDR` - The endpoint of a cluster member (10.253.5.56:5701).
* `HZ_CLUSTERNAME` - The name of the Hazelcast cluster.
* `HZ_USERNAME` - The hazelcast username, should be allowed to destroy maps
* `HZ_PASSWORD` - The password for HZ_USERNAME.
* `DRY_RUN` - Optional, default is true, if set to `false`, it will actually destroy empty maps, otherwise it will only print the maps that would be destroyed.
