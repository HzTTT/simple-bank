#!/bin/sh

#the scrip will exit if a command return non-zero status
set -e

echo "run db migration"
socure /app/app.sh
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"

#take all the parameters passed to the scrip and run it
echo "$@"
exec "$@"