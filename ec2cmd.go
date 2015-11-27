package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"

	"github.com/reiki4040/cstore"
	myec2 "github.com/reiki4040/rnzoo/ec2"
)

const (
	EC2LIST_DESC = `
     you can set default region by AWS_REGION environment variable.

         export AWS_REGION=ap-northeast-1

     this command make cache file that ec2 info. (default ~/.rnzoo/instance.cache.REGION)
     second time, you can get ec2 info without access to AWS.

     if you updated ec2(create new instance, stop, start and etc...), need to update cache with -f/--force option.

         ec2list -r ap-northeast-1 -f`

	EC2LIST_USAGE = `show your ec2 infomations with LTSV format.

     show your ec2 info at ap-northeast-1

       rnzoo ec2list -r ap-northeast-1

       instance_id:i-11111111	name:Name tags Value1	state:stopped	public_ip:X.X.X.X	private_ip:Y.Y.Y.Y	instance_type:t2.micro
       instance_id:i-22222222	name:Name tags Value2	state:running	public_ip:X.X.X.x	private_ip:Y.Y.Y.y	instance_type:m3.large
       ...
`

	EC2LIST_FORCE_USAGE  = `reload ec2 (force connect to AWS)`
	EC2LIST_REGION_USAGE = `specify AWS region name.`
)

var commandEc2list = cli.Command{
	Name:        "ec2list",
	ShortName:   "ls",
	Usage:       EC2LIST_USAGE,
	Description: EC2LIST_DESC,
	Action:      doEc2list,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  OPT_FORCE + ", f",
			Usage: EC2LIST_FORCE_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
	},
}

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

func doEc2list(c *cli.Context) {
	isReload := c.Bool(OPT_FORCE)

	regionName := c.String(OPT_REGION)
	if regionName == "" {
		regionName = os.Getenv(ENV_AWS_REGION)
	}

	if regionName == "" {
		log.Fatalf("please set region.")
	}

	err := CreateRnzooDir()
	if err != nil {
		log.Printf("can not create rnzoo dir: %s\n", err.Error())
	}

	h, err := NewRnzooCStoreManager()
	if err != nil {
		log.Printf("can not load EC2: %s\n", err.Error())
	}

	ec2list, err := h.LoadChoosableEC2List(regionName, isReload)
	if err != nil {
		log.Printf("can not load EC2: %s\n", err.Error())
	}

	for _, i := range ec2list {
		fmt.Println(i.Choice())
	}
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
	var ids []*string
	if instanceId == "" {

		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		ids, err = h.ChooseEC2(region, false)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

	} else {
		ids = []*string{aws.String(instanceId)}
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

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
	var ids []*string
	if instanceId == "" {

		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		ids, err = h.ChooseEC2(region, false)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

	} else {
		ids = []*string{aws.String(instanceId)}
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

func NewRnzooCStoreManager() (*myec2.EC2Handler, error) {
	dirPath := GetRnzooDir()
	m, err := cstore.NewManager("rnzoo", dirPath)
	if err != nil {
		return nil, err
	}

	return myec2.NewEC2Handler(m), nil
}

func GetDefaultRegion() (string, error) {
	region := os.Getenv(ENV_AWS_REGION)
	if region == "" {
		err := fmt.Errorf("does not specify region.")
		return "", err

	}

	return region, nil
}