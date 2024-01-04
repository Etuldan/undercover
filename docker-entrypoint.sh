#!/bin/sh
# docker-entrypoint.sh

sed -i -e "s|SEQ_URL|$SEQ_URL|g" /app/config/config.yml
sed -i -e "s|SEQ_APIKEY|$SEQ_APIKEY|g" /app/config/config.yml

# Run the standard container command.
exec "$@"