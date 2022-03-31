#!/bin/bash

# the requirement body can be any DDL statements without the 'ALTER TABLE <table_name>' prefix
# For example:
# ADD COLUMN dummy INTEGER
# DROP COLUMN dummy
curl -X POST -H "Content-Type:application/sql" -d @- "http://127.0.0.1:8888/api/v1/sources/db02/tables/members_tbl01/ddl" <<EOF
ADD COLUMN dummy INTEGER
EOF
