#!/bin/bash

redis-cli -h "$REDIS_HOST" -p 6379 -a "$REDIS_PASSWORD" FT.CREATE ingest-entity ON JSON PREFIX 1 "ingest-entity:" SCHEMA $.data_type AS data_type TEXT $.state AS state NUMERIC
