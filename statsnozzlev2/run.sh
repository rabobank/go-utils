# d06:
export API_ADDR=https://api.sys.cfd06.rabobank.nl
export CLIENT_ID=<npc client-id>
export CLIENT_SECRET=<npc-secret>

export SHARD_ID=RLP-SHARD
export STORE_IN_DB=true
export DB_FILE=/tmp/statsnozzlev2.db
export PRINT_LOGS=false
export MAX_MESSAGES=50000
#export URI_FILTER='^/api/.*'
./statsnozzlev2
