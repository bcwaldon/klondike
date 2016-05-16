package monitor

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

type AWSLoadBalancerConfig struct {
	Name      string
	Instances []string
}

type AWSManager interface {
	Instances(tags map[string]string) ([]string, error)
	SyncLoadBalancer(cfg AWSLoadBalancerConfig) error
}

func newAWSManager(region string) AWSManager {
	awscfg := aws.NewConfig()
	awscfg = awscfg.WithRegion(region)
	awscfg = awscfg.WithCredentialsChainVerboseErrors(true)

	return &awsManager{
		cfg: awscfg,
	}
}

type awsManager struct {
	cfg *aws.Config
}

func (am *awsManager) Instances(tags map[string]string) ([]string, error) {
	svc := ec2.New(session.New(am.cfg))

	var filters []*ec2.Filter
	for k, v := range tags {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", k)),
			Values: []*string{aws.String(v)},
		})
	}

	req := ec2.DescribeInstancesInput{
		Filters: filters,
	}

	resp, err := svc.DescribeInstances(&req)
	if err != nil {
		return nil, err
	}

	instances := []string{}

	//TODO(bcwaldon): need to handle resp.NextToken
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			instances = append(instances, *inst.InstanceId)
		}
	}

	return instances, nil
}

func (am *awsManager) SyncLoadBalancer(cfg AWSLoadBalancerConfig) error {
	svc := elb.New(session.New(am.cfg))

	req := elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(cfg.Name)},
	}

	resp, err := svc.DescribeLoadBalancers(&req)
	if err != nil {
		return err
	}

	if len(resp.LoadBalancerDescriptions) == 0 {
		return fmt.Errorf("could not find load balancer with name %q", cfg.Name)
	}

	want := map[string]struct{}{}
	for _, instID := range cfg.Instances {
		want[instID] = struct{}{}
	}

	obj := resp.LoadBalancerDescriptions[0]
	got := map[string]struct{}{}
	for _, inst := range obj.Instances {
		got[*inst.InstanceId] = struct{}{}
	}

	var add []string
	var remove []string

	for id := range want {
		if _, ok := got[id]; !ok {
			add = append(add, id)
		}
	}

	for id := range got {
		if _, ok := want[id]; !ok {
			remove = append(remove, id)
		}
	}

	if len(add) > 0 {
		if err := am.addToLoadBalancer(svc, cfg.Name, add); err != nil {
			return err
		}
		log.Printf("Added instances to load balancer %s: %+v", cfg.Name, add)
	}

	if len(remove) > 0 {
		if err := am.removeFromLoadBalancer(svc, cfg.Name, remove); err != nil {
			return err
		}
		log.Printf("Removed instances from load balancer %s: %+v", cfg.Name, remove)
	}

	return nil
}

func (am *awsManager) addToLoadBalancer(svc *elb.ELB, name string, instances []string) error {
	input := make([]*elb.Instance, len(instances))
	for i := range instances {
		input[i] = &elb.Instance{
			InstanceId: aws.String(instances[i]),
		}
	}
	req := elb.RegisterInstancesWithLoadBalancerInput{
		LoadBalancerName: aws.String(name),
		Instances:        input,
	}
	if _, err := svc.RegisterInstancesWithLoadBalancer(&req); err != nil {
		return err
	}
	return nil
}

func (am *awsManager) removeFromLoadBalancer(svc *elb.ELB, name string, instances []string) error {
	input := make([]*elb.Instance, len(instances))
	for i := range instances {
		input[i] = &elb.Instance{
			InstanceId: aws.String(instances[i]),
		}
	}
	req := elb.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: aws.String(name),
		Instances:        input,
	}
	if _, err := svc.DeregisterInstancesFromLoadBalancer(&req); err != nil {
		return err
	}
	return nil
}
