#!/bin/bash -e

if [ -z "$LOGSTASH_HOST" ]; then
	echo "LOGSTASH_HOST is unset"
	exit 2	
fi

cp journalbeat.yml.tmpl journalbeat.yml
cat << EOF >> journalbeat.yml
output:
  logstash:
    hosts: ["${LOGSTASH_HOST}"]
EOF

./journalbeat -e -v
