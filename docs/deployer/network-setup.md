# Network Setup

Deploying a klondike cluster requires that a VPC be provisioned beforehand.
A CloudFormation template exists in contrib/ that may be used to deploy the bare-minimum network, automatically configuring a peering connection with an existing VPC:

```
aws cloudformation create-stack --stack-name klondike-vpc \
  --template-body=file://contrib/network-stack-template.yaml \
  --parameters ParameterKey=ClusterNetworkCIDR,ParameterValue=${ClusterNetworkCIDR} \
               ParameterKey=PeerVPC,ParameterValue=${PeerVPC} \
               ParameterKey=PeerRouteTable,ParameterValue=${PeerRouteTable}
```
