#!/usr/bin/env sh

set -e

echo "Create database"
sqlite3 ./db/database.db < ./db/schema.sql

echo "Set permissions"
chmod a+rw ./db/database.db