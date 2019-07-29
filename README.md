# Go-Libra

## Deploy

### Run API Server

```bash
export GO111MODULE=on
export LIBRA_MYSQL_URL="DATABASE_USERNAME:DATABASE_PASSWORD@(DATABASE_HOST:DATABASE_POSRT)/DATABASE_NAME" 
# example "root:test@(127.0.0.1:3306)/libra"
go build main.go
./main
```

### Run Block Fetcher

```bash
export GO111MODULE=on
export LIBRA_MYSQL_URL="DATABASE_USERNAME:DATABASE_PASSWORD@(DATABASE_HOST:DATABASE_POSRT)/DATABASE_NAME" 
# example "root:test@(127.0.0.1:3306)/libra"

# telegram notify setting
export LIBRA_BOT_KEY=""
export LIBRA_BOT_SECRET=""
export LIBRA_BOT_CHAT_ID=""

go build block_fetcher.go
./block_fetcher
```
