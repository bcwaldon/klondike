#!/usr/bin/env python
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

import argparse
import json

import botocore.session


def tag_dict(tags):
    return dict([(item['Key'], item['Value']) for item in tags])


def _stack_name(cluster):
    return 'klondike-cluster-{}'.format(cluster)


def get_stack_info(cfn_client, cluster):
    stack_name = _stack_name(cluster)
    info = {
        'stack_name': stack_name,
    }

    cfn_resp = cfn_client.describe_stacks(StackName=stack_name)

    info['stack_status'] = cfn_resp['Stacks'][0]['StackStatus']
    return info


def get_bastion_info(cfn_client, cluster):
    res = cfn_client.describe_stack_resource(
        StackName=_stack_name(cluster),
        LogicalResourceId='BastionDNSRecordSet')
    det = res.get('StackResourceDetail', {})

    info = {
        'bastion_host': det.get('PhysicalResourceId'),
    }

    return info


def get_worker_info(asg_client, cluster):
    asg_resp = asg_client.describe_auto_scaling_groups()

    groups = [g for g in asg_resp['AutoScalingGroups']
              if tag_dict(g['Tags']).get('KubernetesCluster') == cluster]

    info = {
        'worker_asg_name': '',
        'worker_count': '',
        'worker_asg_capacity': '',
    }

    worker_group = None
    for g in groups:
        tags = tag_dict(g['Tags'])
        tag_group = tags.get('group')
        if tag_group == 'worker':
            worker_group = g
            break

    if worker_group is None:
        return info

    info['worker_asg_name'] = worker_group['AutoScalingGroupName']

    count = len(worker_group['Instances'])
    info['worker_count'] = count
    desired = worker_group['DesiredCapacity']
    info['worker_asg_capacity'] = desired

    return info


def get_cluster_info(cluster):
    session = botocore.session.get_session()
    cfn_client = session.create_client('cloudformation')
    asg_client = session.create_client('autoscaling')

    info = []

    stack = get_stack_info(cfn_client, cluster)
    info.extend(stack.items())

    if stack['stack_status'] != "CREATE_COMPLETE":
        return info

    info.extend(get_bastion_info(cfn_client, cluster).items())
    info.extend(get_worker_info(asg_client, cluster).items())

    return info


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("cluster", help="klondike cluster name")
    parser.add_argument("--debug", action='store_true',
                        help="print debug-level information")
    parser.add_argument("--json", action='store_true',
                        help="print cluster info as a JSON rather than YAML")
    args = parser.parse_args()

    info = get_cluster_info(args.cluster)

    if args.json:
        print json.dumps(dict(info))
    else:
        for item in info:
            print "{}: {}".format(*item)
