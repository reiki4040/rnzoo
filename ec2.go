package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/codegangsta/cli"
)

var commandEc2start = cli.Command{
	Name:        "ec2start",
	Usage:       "start ec2",
	Description: `start ec2 that already exists.`,
	Action:      doEc2start,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify start instance id.",
		},
	},
}

var commandEc2stop = cli.Command{
	Name:        "ec2stop",
	Usage:       "stop ec2",
	Description: `stop ec2 that already running.`,
	Action:      doEc2stop,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify stop instance id.",
		},
	},
}

func doEc2start(c *cli.Context) {
	prepare(c)

	region := c.String(OPT_REGION)
	if region == "" {
		defaultRegion, err := GetDefaultRegion()
		if err != nil {
			log.Fatalf(err.Error())
		}

		region = defaultRegion
	}

	instanceId := c.String(OPT_INSTANCE_ID)
	ids := make([]*string, 0, 1)
	if instanceId == "" {
		instanceIds, err := ChooseEC2(region, "select instance")
		if err != nil {
			log.Fatalf(err.Error())
			return
		}

		for _, i := range instanceIds {
			ids = append(ids, aws.String(i))
		}
	} else {
		ids = append(ids, aws.String(instanceId))
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})

	params := &ec2.StartInstancesInput{
		InstanceIds: ids,
	}

	resp, err := cli.StartInstances(params)
	if err != nil {
		log.Fatalf("error during launching: %s", err.Error())
		return
	}

	for _, status := range resp.StartingInstances {
		id := convertNilString(status.InstanceId)
		pState := convertNilString(status.PreviousState.Name)
		cState := convertNilString(status.CurrentState.Name)
		log.Printf("launched %s: %s -> %s", id, pState, cState)
	}

	log.Printf("finished launching.")
}

func doEc2stop(c *cli.Context) {
	prepare(c)

	region := c.String(OPT_REGION)
	if region == "" {
		defaultRegion, err := GetDefaultRegion()
		if err != nil {
			log.Fatalf(err.Error())
		}
		region = defaultRegion
	}

	instanceId := c.String(OPT_INSTANCE_ID)
	ids := make([]*string, 0, 1)
	if instanceId == "" {
		instanceIds, err := ChooseEC2(region, "select instance")
		if err != nil {
			log.Fatalf(err.Error())
			return
		}

		for _, i := range instanceIds {
			ids = append(ids, aws.String(i))
		}
	} else {
		ids = append(ids, aws.String(instanceId))
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})
	params := &ec2.StopInstancesInput{
		InstanceIds: ids,
	}

	resp, err := cli.StopInstances(params)
	if err != nil {
		log.Fatalf("error during stopping: %s", err.Error())
		return
	}

	for _, status := range resp.StoppingInstances {
		id := convertNilString(status.InstanceId)
		pState := convertNilString(status.PreviousState.Name)
		cState := convertNilString(status.CurrentState.Name)
		log.Printf("stopped %s: %s -> %s", id, pState, cState)
	}

	log.Printf("finished stopping.")
}

func GetDefaultRegion() (string, error) {
	region := os.Getenv(ENV_AWS_REGION)
	if region == "" {
		err := fmt.Errorf("does not specify region.")
		return "", err

	}

	return region, nil
}
