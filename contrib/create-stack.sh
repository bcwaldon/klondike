#!/bin/bash -ex

source ./contrib/stackrc

vars=(
	"TEMPLATE_URL"
	"CLUSTER_NAME"
	"STACK_NAME"
	"AZ"
	"CONTROLLER_INSTANCE_TYPE"
	"WORKER_INSTANCE_TYPE"
	"SSH_KEYNAME"
	"IMAGE_ID"
	"USER_DATA_FILE"
)

for var in "${vars[@]}"; do
	eval val=\$$var
	if [ -z "$val" ]; then
		echo "Must set $var"
		exit 1;
	fi
done

USER_DATA=$(cat $USER_DATA_FILE | base64)

PARAMETERS=( 
	"ParameterKey=AvailabilityZone,ParameterValue=$AZ"
	"ParameterKey=ControllerInstanceType,ParameterValue=$CONTROLLER_INSTANCE_TYPE"
	"ParameterKey=WorkerInstanceType,ParameterValue=$WORKER_INSTANCE_TYPE"
	"ParameterKey=SSHKeyName,ParameterValue=$SSH_KEYNAME"
	"ParameterKey=CoreOSImageID,ParameterValue=$IMAGE_ID"
	"ParameterKey=ClusterName,ParameterValue=$CLUSTER_NAME"
	"ParameterKey=ControllerUserData,ParameterValue=$USER_DATA"
	"ParameterKey=WorkerUserData,ParameterValue=$USER_DATA"
)

function join { local IFS="$1"; shift; echo "$*"; }

aws cloudformation create-stack \
	--stack-name $STACK_NAME \
	--capabilities CAPABILITY_IAM \
	--parameters $(join " " "${PARAMETERS[@]}") \
	--template-url $TEMPLATE_URL \
	--disable-rollback
