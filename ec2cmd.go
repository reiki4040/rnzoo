package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"

	"github.com/reiki4040/cstore"
	myec2 "github.com/reiki4040/rnzoo/ec2"
)

const (
	EC2LIST_DESC = `
     show your ec2 info at ap-northeast-1

       rnzoo ec2list -r ap-northeast-1

       i-11111111	Name tag Web server1	stopped	t2.micro	54.X.X.X	10.Y.Y.Y
       i-22222222	Name tag Web server2	running	m3.large	52.X.X.x	10.Y.Y.y
       ...

     you can set default region by AWS_REGION environment variable.

         export AWS_REGION=ap-northeast-1

     this command make cache file that ec2 info. (default ~/.rnzoo/aws.instance.cache.REGION)
     second time, you can get ec2 info without access to AWS.

     if you updated ec2(create new instance, stop, start and etc...), need to update cache with -f/--force option.

         ec2list -r ap-northeast-1 -f`

	EC2LIST_USAGE = `show your ec2 instances.`

	EC2LIST_FORCE_USAGE  = `reload ec2 (force connect to AWS)`
	EC2LIST_REGION_USAGE = `specify AWS region name.`
)

var commandInit = cli.Command{
	Name:        "init",
	Usage:       "initialize settings",
	Description: `start initialize settings wizard`,
	Action:      doInit,
	Flags:       []cli.Flag{},
}

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

	region := c.String(OPT_REGION)
	if region == "" {
		// load config
		c, err := GetDefaultConfig()
		if err != nil {
			log.Printf("can not load rnzoo config: %s\n", err.Error())
		}

		region = c.AWSRegion
	}

	if region == "" {
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

	ec2list, err := h.LoadChoosableEC2List(region, isReload)
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
		// load config
		c, err := GetDefaultConfig()
		if err != nil {
			log.Printf("can not load rnzoo config: %s\n", err.Error())
		}

		region = c.AWSRegion
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
		// load config
		c, err := GetDefaultConfig()
		if err != nil {
			log.Printf("can not load rnzoo config: %s\n", err.Error())
		}

		region = c.AWSRegion
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

var commandAttachEIP = cli.Command{
	Name:        "attach-eip",
	Usage:       "allocate new EIP(allow reassociate) and associate it to the instance.",
	Description: `allocate new EIP(allow reassociate) and associate it to the instance.`,
	Action:      doAttachEIP,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify instance id.",
		},
		cli.BoolFlag{
			Name:  OPT_REUSE,
			Usage: "if there is EIP that has not associated, associate it. if not, allocate new address.",
		},
	},
}

var commandDetachEIP = cli.Command{
	Name:        "detach-eip",
	Usage:       "disassociate EIP and release it.",
	Description: `disassociate EIP and release it.`,
	Action:      doDetachEIP,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify instance id.",
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_RELEASE,
			Usage: "does not release disassociated the address.",
		},
	},
}

// allocate new EIP and associate.
func doAttachEIP(c *cli.Context) {
	prepare(c)

	// load config
	config, err := GetDefaultConfig()
	if err != nil {
		log.Printf("can not load rnzoo config: %s\n", err.Error())
	}
	region := config.AWSRegion

	instanceId := c.String(OPT_INSTANCE_ID)
	if instanceId == "" {
		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		ids, err := h.ChooseEC2(region, true)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

		// one instance
		if len(ids) >= 1 {
			instanceId = *ids[0]
		}
	} else {
		err := validateInstanceId(instanceId)
		if err != nil {
			log.Fatalf("invalid instance id format: %s", err.Error())
		}
	}

	reuseEIP := c.Bool(OPT_REUSE)

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	var allocId string
	var ip string
	if reuseEIP {
		address, err := myec2.GetNotAssociateEIP(cli)
		if err != nil {
			log.Printf("failed no associate address so allocate new address...")
		}

		// if exists EIP
		if address != nil {
			allocId = convertNilString(address.AllocationId)
			ip = convertNilString(address.PublicIp)
		}
	}

	if allocId == "" {
		aid, pip, err := myec2.AllocateEIP(cli)
		if err != nil {
			log.Fatalf("failed allocation address:%s", err.Error())
		}
		allocId = convertNilString(aid)
		ip = convertNilString(pip)

		log.Printf("allocated allocation_id:%s\tpublic_ip:%s", allocId, ip)
	}

	associationId, err := myec2.AssociateEIP(cli, allocId, instanceId)
	if err != nil {
		log.Fatalf("failed associate address:%s", err.Error())
	}

	log.Printf("associated association_id:%s\tpublic_ip:%s\tinstance_id:%s", convertNilString(associationId), ip, instanceId)
}

// desassociate EIP and release.
func doDetachEIP(c *cli.Context) {
	prepare(c)

	withoutRelease := c.Bool(OPT_WITHOUT_RELEASE)

	// load config
	config, err := GetDefaultConfig()
	if err != nil {
		log.Printf("can not load rnzoo config: %s\n", err.Error())
	}

	region := config.AWSRegion

	instanceId := c.String(OPT_INSTANCE_ID)
	if instanceId == "" {
		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		ids, err := h.ChooseEC2(region, true)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

		// one instance
		if len(ids) >= 1 {
			instanceId = *ids[0]
		}
	} else {
		err := validateInstanceId(instanceId)
		if err != nil {
			log.Fatalf("invalid instance id format: %s", err.Error())
		}
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})
	address, err := myec2.GetEIPFromInstance(cli, instanceId)
	if err != nil {
		log.Fatalf(err.Error())
	}

	associationId := convertNilString(address.AssociationId)
	ip := convertNilString(address.PublicIp)
	iid := convertNilString(address.InstanceId)

	err = myec2.DisassociateEIP(cli, convertNilString(address.AssociationId))
	if err != nil {
		log.Fatalf("failed disassociate address:%s", err.Error())
	}

	log.Printf("disassociated assciation_id:%s\tpublic_ip:%s\tinstance_id:%s", associationId, ip, iid)

	if !withoutRelease {
		err := myec2.ReleaseEIP(cli, convertNilString(address.AllocationId))
		if err != nil {
			log.Fatalf("failed release address:%s", err.Error())
		}
		log.Printf("released allocation_id:%s\tpublic_ip:%s", convertNilString(address.AllocationId), ip)
	}
}

func NewCStoreManager() (*cstore.Manager, error) {
	dirPath := GetRnzooDir()
	return cstore.NewManager("rnzoo", dirPath)
}

func NewRnzooCStoreManager() (*myec2.EC2Handler, error) {
	m, err := NewCStoreManager()
	if err != nil {
		return nil, err
	}

	return myec2.NewEC2Handler(m), nil
}
