#!/bin/sh
#
# Copyright 2019 PingCAP, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

set -eu

DB="$TEST_NAME"
TABLE="usertable"

run_sql "CREATE DATABASE $DB;"

go-ycsb load mysql -P tests/$TEST_NAME/workload -p mysql.host=$TIDB_IP -p mysql.port=$TIDB_PORT -p mysql.user=root -p mysql.db=$DB

row_count_ori=$(run_sql "SELECT COUNT(*) FROM $DB.$TABLE;" | awk '/COUNT/{print $2}')

# Update GC safepoint to now + 5s after 10s seconds.
sleep 10 && bin/gc -pd $PD_ADDR -gc-offset "5s" &

# Set ratelimit to 1 bytes/second, we assume it can not finish within 10s,
# so it will trigger exceed GC safe point error.
backup_gc_fail=0
echo "backup start (expect fail)..."
run_br --pd $PD_ADDR backup table -s "local://$TEST_DIR/$DB" --db $DB -t $TABLE --ratelimit 1 --ratelimit-unit 1 || backup_gc_fail=1

if [ "$backup_gc_fail" -ne "1" ];then
    echo "TEST: [$TEST_NAME] failed!"
    exit 1
fi

run_sql "DROP TABLE $DB.$TABLE;"
