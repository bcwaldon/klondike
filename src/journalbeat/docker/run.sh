#
# Copyright 2016 Planet Labs
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
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
