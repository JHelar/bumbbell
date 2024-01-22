#!/bin/bash

set -e

cd db

echo "Create database"
sqlite-utils create-database database.db

echo "Create tables"
sqlite3 database.db < schema.sql

echo "Seed users table"
sqlite-utils insert database.db users seed/users.json --truncate

echo "Seed splits table"
sqlite-utils insert database.db splits seed/splits.json --truncate

echo "Seed images table"
sqlite-utils insert database.db images seed/images.json --truncate

echo "Seed exercises table"
sqlite-utils insert database.db exercises seed/exercises.json --truncate