package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"os"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create VPC
		vpc, err := ec2.NewVpc(ctx, "myVpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.10.0.0/16"),
			EnableDnsHostnames: pulumi.Bool(true),
			EnableDnsSupport:   pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("MyVPC"),
			},
		})
		if err != nil {
			return err
		}

		// Create Public Subnet
		publicSubnet, err := ec2.NewSubnet(ctx, "publicSubnet", &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String("10.10.1.0/24"),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			AvailabilityZone:    pulumi.String("ap-southeast-1a"),
		})
		if err != nil {
			return err
		}

		// Create Internet Gateway
		igw, err := ec2.NewInternetGateway(ctx, "igw", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
		})
		if err != nil {
			return err
		}

		// Create Route Table
		routeTable, err := ec2.NewRouteTable(ctx, "routeTable", &ec2.RouteTableArgs{
			VpcId: vpc.ID(),
			Routes: ec2.RouteTableRouteArray{
				&ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: igw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}

		// Associate Route Table with Public Subnet
		_, err = ec2.NewRouteTableAssociation(ctx, "rtAssocPublic", &ec2.RouteTableAssociationArgs{
			SubnetId:     publicSubnet.ID(),
			RouteTableId: routeTable.ID(),
		})
		if err != nil {
			return err
		}

		// Create Security Group
		securityGroup, err := ec2.NewSecurityGroup(ctx, "webSecGrp", &ec2.SecurityGroupArgs{
			Description: pulumi.String("Enable SSH and K3s access"),
			VpcId:       vpc.ID(),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(22),
					ToPort:     pulumi.Int(22),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
				&ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(6443),
					ToPort:     pulumi.Int(6443),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		})
		if err != nil {
			return err
		}

		// Get the public key from environment variable
		publicKey := os.Getenv("PUBLIC_KEY")

		// Create EC2 KeyPair
		keyPair, err := ec2.NewKeyPair(ctx, "myKeyPair", &ec2.KeyPairArgs{
			KeyName:   pulumi.String("my-key-pair"),
			PublicKey: pulumi.String(publicKey),
		})
		if err != nil {
			return err
		}

		amiID := "ami-003c463c8207b4dfa" // Replace with a valid AMI ID for your region
		instanceType := "t3.small"

		// Create Master Instances
		masterNode, err := ec2.NewInstance(ctx, "masterNode", &ec2.InstanceArgs{
			InstanceType:        pulumi.String(instanceType),
			Ami:                 pulumi.String(amiID),
			SubnetId:            publicSubnet.ID(),
			KeyName:             keyPair.KeyName,
			VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
			Tags:                pulumi.StringMap{"Name": pulumi.String("master-node")},
		})
		if err != nil {
			return err
		}

		// Create Worker1 Instances
		workerNode1, err := ec2.NewInstance(ctx, "workerNode1", &ec2.InstanceArgs{
			InstanceType:        pulumi.String(instanceType),
			Ami:                 pulumi.String(amiID),
			SubnetId:            publicSubnet.ID(),
			KeyName:             keyPair.KeyName,
			VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
			Tags:                pulumi.StringMap{"Name": pulumi.String("worker-node-1")},
		})
		if err != nil {
			return err
		}

		// Create Worker2 Instances
		workerNode2, err := ec2.NewInstance(ctx, "workerNode2", &ec2.InstanceArgs{
			InstanceType:        pulumi.String(instanceType),
			Ami:                 pulumi.String(amiID),
			SubnetId:            publicSubnet.ID(),
			KeyName:             keyPair.KeyName,
			VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
			Tags:                pulumi.StringMap{"Name": pulumi.String("worker-node-2")},
		})
		if err != nil {
			return err
		}

		// Create Nginx Instances
		nginxInstance, err := ec2.NewInstance(ctx, "nginxInstance", &ec2.InstanceArgs{
			InstanceType:        pulumi.String(instanceType),
			Ami:                 pulumi.String(amiID),
			SubnetId:            publicSubnet.ID(),
			KeyName:             keyPair.KeyName,
			VpcSecurityGroupIds: pulumi.StringArray{securityGroup.ID()},
			Tags:                pulumi.StringMap{"Name": pulumi.String("nginx-instance")},
		})
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("masterPublicIp", masterNode.PublicIp)
		ctx.Export("worker1PublicIp", workerNode1.PublicIp)
		ctx.Export("worker2PublicIp", workerNode2.PublicIp)
		ctx.Export("nginxInstanceIp", nginxInstance.PublicIp)

		return nil
	})
}
