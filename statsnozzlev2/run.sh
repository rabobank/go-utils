export API_ADDR=https://api.sys.mydomain.com
export CLIENT_ID=statsnozzlev2-client
export CLIENT_SECRET=my-client-secret

export SHARD_ID=RLP-SHARD
export STORE_IN_DB=true
export DB_FILE=/tmp/statsnozzlev2.db
export PRINT_LOGS=true
export MAX_MESSAGES=50000
export URI_FILTER='^/api/.*'
./statsnozzlev2
