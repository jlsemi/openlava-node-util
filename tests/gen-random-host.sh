#!/bin/bash
export SQLITE_DB_PATH=${PWD}/openlava.db
NODECLI="../build/nodecli-0.2.0"

rm -rf ${PWD}/openlava.db
for i in {1..10}; do
	hostname="ehpc-test-$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c10)"
	${NODECLI} add --config_dir=/tmp --hostname=${hostname}
done

