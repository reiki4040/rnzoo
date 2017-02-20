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
	"github.com/reiki4040/peco"
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
	EC2LIST_TSV          = `output with tab separate format (TSV)`

	EC2TYPE_DESC = `
	modify EC2 isntacne type. the instance must be already stopped.
	the max of type in selection list are t2, c4, m4, r4 series's large size.
	if you want other types, please use -t, --type option.`
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
		cli.BoolFlag{
			Name:  OPT_TSV + ", t",
			Usage: EC2LIST_TSV,
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

var commandEc2type = cli.Command{
	Name:        "ec2type",
	Usage:       "modify ec2 isntance type",
	Description: EC2TYPE_DESC,
	Action:      doEc2type,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify already stopped instance id.",
		},
		cli.StringFlag{
			Name:  OPT_I_TYPE + ", t",
			Usage: "specify new instance type.",
		},
		cli.BoolFlag{
			Name:  OPT_START,
			Usage: "start the instance after modifying type.",
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

	ec2list, err := h.LoadChoosableEC2List(region, myec2.EC2_STATE_ANY, isReload)
	if err != nil {
		log.Printf("can not load EC2: %s\n", err.Error())
	}

	for _, i := range ec2list {
		if c.Bool(OPT_TSV) {
			fmt.Println(i)
		} else {
			fmt.Println(i.Choice())
		}
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

		ids, err = h.ChooseEC2(region, myec2.EC2_STATE_STOPPED, true)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

	} else {
		ids = []*string{aws.String(instanceId)}
	}

	resp, err := startInstances(region, ids)
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

func startInstances(region string, ids []*string) (*ec2.StartInstancesOutput, error) {
	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	params := &ec2.StartInstancesInput{
		InstanceIds: ids,
	}

	return cli.StartInstances(params)
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

		ids, err = h.ChooseEC2(region, myec2.EC2_STATE_RUNNING, true)
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

func doEc2type(c *cli.Context) {
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

		ids, err = h.ChooseEC2(region, myec2.EC2_STATE_STOPPED, true)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

	} else {
		ids = []*string{aws.String(instanceId)}
	}

	iType := c.String(OPT_I_TYPE)
	if iType == "" {
		chosenType, err := peco.Choose("Instance Type", "Please select Instance Type", "", EC2InstanceTypeList)
		if err != nil {
			log.Fatalf("error during select instance type: %s", err.Error())
			return
		}

		if len(chosenType) != 1 {
			log.Fatal("multiple type selected. please single type.")
			return
		}

		iType = chosenType[0].Value()
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	for _, i := range ids {
		params := &ec2.ModifyInstanceAttributeInput{
			InstanceId: i,
			InstanceType: &ec2.AttributeValue{
				Value: aws.String(iType),
			},
		}

		// resp is empty
		_, err := cli.ModifyInstanceAttribute(params)
		if err != nil {
			log.Fatalf("error during modify instance type: %s", err.Error())
			return
		}

		log.Printf("%s is modified the instance type to %s", *i, iType)

		if c.Bool(OPT_START) {
			resp, err := startInstances(region, []*string{i})
			if err != nil {
				log.Fatalf("error during starting instance: %s", err.Error())
				return
			}

			for _, status := range resp.StartingInstances {
				id := convertNilString(status.InstanceId)
				pState := convertNilString(status.PreviousState.Name)
				cState := convertNilString(status.CurrentState.Name)
				log.Printf("launched %s: %s -> %s", id, pState, cState)
			}
		}
	}

	log.Printf("finished modifying instance type.")
}

var (
	EC2InstanceTypeList = []peco.Choosable{
		&peco.Choice{C: "t2.nano", V: "t2.nano"},
		&peco.Choice{C: "t2.micro", V: "t2.micro"},
		&peco.Choice{C: "t2.small", V: "t2.small"},
		&peco.Choice{C: "t2.medium", V: "t2.medium"},
		&peco.Choice{C: "t2.large", V: "t2.large"},
		&peco.Choice{C: "c4.large", V: "c4.large"},
		&peco.Choice{C: "m4.large", V: "m4.large"},
		&peco.Choice{C: "r4.large", V: "r4.large"},
	}
)

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
		cli.BoolFlag{
			Name:  OPT_MOVE,
			Usage: "move EIP to the instance from the other instance.",
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

func doMoveEIP(region string) {
	// EIP listing
	allocIds, err := myec2.ChooseEIP(region)

	if len(allocIds) == 0 {
		log.Fatalf("error during selecting to EIP: %s", err.Error())
	}

	// to instance
	h, err := NewRnzooCStoreManager()
	if err != nil {
		log.Printf("can not load EC2: %s\n", err.Error())
	}

	ids, err := h.ChooseEC2(region, myec2.EC2_STATE_ANY, true)
	if err != nil {
		log.Fatalf("error during selecting: %s", err.Error())
		return
	}

	// one instance
	instanceId := ""
	if len(ids) >= 1 {
		instanceId = *ids[0]
	} else {
		log.Fatalf("error during selecting to instance: %s", err.Error())
	}

	// moving
	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	assocId, err := myec2.AssociateEIP(cli, allocIds[0].AllocationId, instanceId)
	if err != nil {
		log.Fatalf("error during moving EIP: %s", err.Error())
	}

	log.Printf("associated association_id:%s\tpublic_ip:%s\tinstance_id:%s", convertNilString(assocId), "EIP", instanceId)
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

	moving := c.Bool(OPT_MOVE)
	if moving {
		doMoveEIP(region)
		os.Exit(0)
	}

	instanceId := c.String(OPT_INSTANCE_ID)
	if instanceId == "" {
		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		ids, err := h.ChooseEC2(region, myec2.EC2_STATE_ANY, true)
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

		ids, err := h.ChooseEC2(region, myec2.EC2_STATE_ANY, true)
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
