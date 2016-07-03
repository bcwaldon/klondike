import argparse
import json

import botocore.session


def tag_dict(tags):
    return dict([(item['Key'], item['Value']) for item in tags])


def get_stack_info(cfn_client, cluster):
    stack_name = 'klondike-cluster-{}'.format(cluster)
    info = {
        'stack_name': stack_name,
    }

    cfn_resp = cfn_client.describe_stacks(StackName=stack_name)

    info['stack_status'] = cfn_resp['Stacks'][0]['StackStatus']
    return info


def get_bastion_info(asg_client, ec2_client, cluster):
    asg_resp = asg_client.describe_auto_scaling_groups()

    groups = [g for g in asg_resp['AutoScalingGroups']
              if tag_dict(g['Tags']).get('KubernetesCluster') == cluster]

    info = {
        'bastion_public_ip': '',
    }

    bastion_group = None
    for g in groups:
        tags = tag_dict(g['Tags'])
        tag_group = tags.get('group')
        if tag_group == 'bastion':
            bastion_group = g
            break

    if bastion_group is None:
        return info

    instance_id = bastion_group['Instances'][0]['InstanceId']
    ec2_resp = ec2_client.describe_instances(InstanceIds=[instance_id])
    instance = ec2_resp['Reservations'][0]['Instances'][0]
    info['bastion_public_ip'] = instance['PrivateIpAddress']

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
    ec2_client = session.create_client('ec2')

    info = []

    stack = get_stack_info(cfn_client, cluster)
    info.extend(stack.items())

    if stack['stack_status'] != "CREATE_COMPLETE":
        return info

    info.extend(get_bastion_info(asg_client, ec2_client, cluster).items())
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
