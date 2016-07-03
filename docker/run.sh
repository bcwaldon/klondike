#!/bin/bash -e

cp journalctl.yml.tmpl journalctl.yml
cat << EOF >> journalctl.yml
output:
  logstash:
    hosts: ["${LOGSTASH_HOST}"]
EOF

./journalbeat -e -v
