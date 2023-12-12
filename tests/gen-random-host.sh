#!/bin/bash
export SQLITE_DB_PATH=${PWD}/openlava.db
NODECLI="../build/nodecli-0.3.1"

rm -rf ${PWD}/openlava.db
for i in {1..10}; do
	hostname="ehpc-test-$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c10)"
	${NODECLI} add --config_dir=/tmp --hostname=${hostname} --queuename=normal
done

