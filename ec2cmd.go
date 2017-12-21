package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

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

	   i-11111111	Name tag Web server1	stopped	t2.micro	54.X.X.X	10.Y.Y.Y	xxxx:xxxx::xxxx
	   i-22222222	Name tag Web server2	running	m3.large	52.X.X.x	10.Y.Y.y	xxxx:xxxx::yyyy
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
	modify EC2 instacne type. the instance must be already stopped.
	the max of type in selection list are t2, c4, m4, r4 series's large size.
	if you want other types, please use -t, --type option.`

	EC2RUN_DESC = `
	run EC2 instances with configuration yaml file.
	`
	EC2TERMINATE_DESC = `
	terminate EC2 instances.

	IMPORTANT: default action is dry run, please set --execute option when do termination.

	default listing instances are only stopped instances.
	if you want select in all state instances, please use --ec2-any-state option.
	`
	DEFAULT_OUTPUT_TEMPLATE = "{{.InstanceId}}\t{{.Name}}\t{{.PublicIp}}\t{{.PrivateIp}}"
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
	ShortName:   "start",
	Usage:       "start ec2",
	Description: `start ec2 that already exists.`,
	Action:      doEc2start,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify start instance id.",
		},
		cli.BoolFlag{
			Name:  OPT_CONFIRM,
			Usage: "confirm target instances before action.",
		},
	},
}

var commandEc2stop = cli.Command{
	Name:        "ec2stop",
	ShortName:   "stop",
	Usage:       "stop ec2",
	Description: `stop ec2 that already running.`,
	Action:      doEc2stop,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify stop instance id.",
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_CONFIRM,
			Usage: "without target instance confirming (default action is do confirming)",
		},
	},
}

var commandEc2type = cli.Command{
	Name:        "ec2type",
	ShortName:   "type",
	Usage:       "modify ec2 instance type",
	Description: EC2TYPE_DESC,
	Action:      doEc2type,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
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
		cli.BoolFlag{
			Name:  OPT_CONFIRM,
			Usage: "confirm target instances before action.",
		},
	},
}

var commandEc2run = cli.Command{
	Name:        "ec2run",
	ShortName:   "run",
	Usage:       "run new ec2 instances",
	Description: EC2RUN_DESC,
	Action:      doEc2run,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  OPT_DRYRUN,
			Usage: "dry-run ec2 run.",
		},
		cli.StringFlag{
			Name:  OPT_SKELETON,
			Usage: "store skeleton config yaml to specified file path",
		},
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_AMI_ID,
			Usage: "overwrite run AMI ID.",
		},
		cli.StringFlag{
			Name:  OPT_I_TYPE,
			Usage: "overwrite run instance type.",
		},
		cli.StringFlag{
			Name:  OPT_SYMBOL,
			Usage: "replace {{.Symbol}} in name tag",
		},
		cli.StringFlag{
			Name:  OPT_SPECIFY_NAME,
			Usage: "specify config name in yaml",
		},
	},
}

var commandEc2terminate = cli.Command{
	Name:        "ec2terminate",
	ShortName:   "terminate",
	Usage:       "terminate instances.",
	Description: EC2TERMINATE_DESC,
	Action:      doEc2Terminate,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify the instance id that you want termination.",
		},
		cli.BoolFlag{
			Name:  OPT_DRYRUN,
			Usage: "dry-run ec2 terminate.",
		},
		cli.BoolFlag{
			Name:  OPT_EXECUTE,
			Usage: "execute ec2 terminate (default action is dryrun. if execute and dryrun options set in same time, then do dryrun)",
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_CONFIRM,
			Usage: "without target instance confirming (default action is do confirming)",
		},
		cli.BoolFlag{
			Name:  OPT_EC2_ANY_STATE,
			Usage: "selectable all state instances (default only stopped instances)",
		},
	},
}

func getRegion(c *cli.Context) (string, error) {
	region := c.String(OPT_REGION)
	if region != "" {
		return region, nil
	}

	region = os.Getenv(ENV_AWS_REGION)
	if region != "" {
		return region, nil
	}

	// load config
	config, err := GetDefaultConfig()
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("can not load rnzoo config: %v", err)
		}
	} else {
		if config.AWSRegion != "" {
			return config.AWSRegion, nil
		}
	}

	return "", fmt.Errorf("did not specified region, please set region with -r option or AWS_REGION environment variable or 'rnzoo init'")
}

func doEc2list(c *cli.Context) {
	isReload := c.Bool(OPT_FORCE)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

	err = CreateRnzooDir()
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

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
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

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	if c.Bool(OPT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, ids...)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		for _, ins := range insts {
			name := "[no Name tag instance]"
			for _, t := range ins.Tags {
				if convertNilString(t.Key) == "Name" {
					name = convertNilString(t.Value)
					break
				}
			}

			fmt.Printf("%s\t%s\t%s\n", convertNilString(ins.InstanceId), name, convertNilString(ins.PrivateIpAddress))
		}

		ans, err := confirm("start above instances?", false)
		if !ans {
			log.Fatalln("canceled instance start action.")
			return
		}
	}

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
}

func doEc2stop(c *cli.Context) {
	prepare(c)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
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

	if !c.Bool(OPT_WITHOUT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, ids...)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		for _, ins := range insts {
			name := "[no Name tag instance]"
			for _, t := range ins.Tags {
				if convertNilString(t.Key) == "Name" {
					name = convertNilString(t.Value)
					break
				}
			}

			fmt.Printf("%s\t%s\t%s\n", convertNilString(ins.InstanceId), name, convertNilString(ins.PrivateIpAddress))
		}

		ans, err := confirm("stop above instances?", false)
		if !ans {
			log.Fatalln("canceled instance stop action.")
			return
		}
	}

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
}

func doEc2type(c *cli.Context) {
	prepare(c)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
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

	if c.Bool(OPT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, ids...)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		for _, ins := range insts {
			name := "[no Name tag instance]"
			for _, t := range ins.Tags {
				if convertNilString(t.Key) == "Name" {
					name = convertNilString(t.Value)
					break
				}
			}

			fmt.Printf("%s\t%s\t%s\t%s\n", convertNilString(ins.InstanceId), name, convertNilString(ins.InstanceType), convertNilString(ins.PrivateIpAddress))
		}

		ans, err := confirm("modified above instance type to "+iType+"?", false)
		if !ans {
			log.Fatalln("canceled instance type change action.")
			return
		}
	}

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
			params := &ec2.StartInstancesInput{
				InstanceIds: []*string{i},
			}

			resp, err := cli.StartInstances(params)
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
}

type EC2RunConfig struct {
	Name               string `yaml:"name"`
	AmiId              string `yaml:"ami_id"`
	IamRoleName        string `yaml:"iam_role_name"`
	PlacementGroupName string `yaml:"placement_group_name" `
	PublicIpEnabled    bool   `yaml:"public_ip_enabled"`
	Ipv6Enabled        bool   `yaml:"ipv6_enabled"`
	Type               string `yaml:"instance_type"`
	KeyPair            string `yaml:"key_pair"`

	EbsDevices   []EC2RunEbs `yaml:"ebs_volumes"`
	EbsOptimized bool        `yaml:"ebs_optimized"`

	Tags             []EC2RunConfigTag    `yaml:"tags"`
	SecurityGroupIds []string             `yaml:"security_group_ids"`
	Launches         []EC2RunConfigLaunch `yaml:"launches"`
}

type EC2RunEbs struct {
	DeviceName          string `yaml:"device_name"`
	DeleteOnTermination bool   `yaml:"delete_on_termination"`
	Encrypted           *bool  `yaml:"encrypted"`
	SizeGB              int64  `yaml:"size_gb"`
	VolumeType          string `yaml:"volume_type"`
}

type EC2RunConfigTag struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}
type EC2RunConfigLaunch struct {
	NameTagTemplate string `yaml:"name_tag_template"`
	SubnetId        string `yaml:"subnet_id"`
	OutputTemplate  string `yaml:"output_template"`
	OverWriteType   string `yaml:"instance_type"`
}

func (c *EC2RunConfig) genLauncher() *myec2.Launcher {
	sgIds := make([]*string, 0, len(c.SecurityGroupIds))
	for _, sgId := range c.SecurityGroupIds {
		sgIds = append(sgIds, aws.String(sgId))
	}

	var roleName *string
	if c.IamRoleName != "" {
		roleName = &c.IamRoleName
	}

	ebss := make([]myec2.Ebs, 0, len(c.EbsDevices))
	for _, e := range c.EbsDevices {
		ebs := myec2.Ebs{
			DeviceName:          e.DeviceName,
			DeleteOnTermination: e.DeleteOnTermination,
			Encrypted:           e.Encrypted,
			SizeGB:              e.SizeGB,
			VolumeType:          e.VolumeType,
		}

		ebss = append(ebss, ebs)
	}

	l := &myec2.Launcher{
		AmiId:              c.AmiId,
		InstanceType:       c.Type,
		KeyName:            c.KeyPair,
		SecurityGroupIds:   sgIds,
		PublicIpEnabled:    c.PublicIpEnabled,
		Ipv6Enabled:        c.Ipv6Enabled,
		IamRoleName:        roleName,
		EbsDevices:         ebss,
		EbsOptimized:       c.EbsOptimized,
		PlacementGroupName: c.PlacementGroupName,
	}

	return l
}

type NameTagReplacement struct {
	Symbol   string
	Sequence string
}

func (r *NameTagReplacement) StringWithTemplate(templateString string) (string, error) {
	t := template.New("instance name template")
	t, err := t.Parse(templateString)
	if err != nil {
		return "", err
	}

	b := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(b)
	err = t.Execute(buf, r)
	if err != nil {
		return "", err
	}

	replacedNameTag := buf.String()
	return replacedNameTag, nil
}

type EC2RunOutput struct {
	InstanceId string
	Name       string
	PrivateIp  string
	PublicIp   string

	Symbol   string
	Sequence string
}

func (o *EC2RunOutput) StringWithTemplate(templateString string) (string, error) {
	t := template.New("ec2run output template")
	t, err := t.Parse(templateString)
	if err != nil {
		return "", err
	}

	b := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(b)
	err = t.Execute(buf, o)
	if err != nil {
		return "", err
	}

	replaced := buf.String()
	return replaced, nil
}

func StoreSkeletonEC2RunConfigYaml(filePath string) error {
	encrypted := true
	s := EC2RunConfig{
		Name:               "skeleton config example. please replace  properties for your case.",
		AmiId:              "ami-xxxxxxx",
		IamRoleName:        "your_iam_role_name",
		PlacementGroupName: "your_exists_placment_group",
		PublicIpEnabled:    false,
		Ipv6Enabled:        false,
		Type:               "t2.nano",
		KeyPair:            "your_key_pair_name",
		EbsOptimized:       false,
		EbsDevices: []EC2RunEbs{
			EC2RunEbs{
				DeviceName:          "/dev/sdb",
				DeleteOnTermination: false,
				Encrypted:           &encrypted,
				SizeGB:              8,
				VolumeType:          "gp2",
			},
		},
		Tags: []EC2RunConfigTag{
			EC2RunConfigTag{
				Key:   "sample_key",
				Value: "sample_value",
			},
		},
		SecurityGroupIds: []string{"sg-xxxxxxxx", "sg-yyyyyyyy"},
		Launches: []EC2RunConfigLaunch{
			EC2RunConfigLaunch{
				NameTagTemplate: "instance {{.Symbol}} {{.Sequence}}",
				SubnetId:        "subnet-xxxxxxxx",
				OutputTemplate:  "{{.InstanceId}},{{.Name}},{{.PublicIp}},{{.Symbol}},{{.Sequence}}",
			},
		},
	}

	return cstore.StoreToYamlFile(filePath, []EC2RunConfig{s})
}

func doEc2run(c *cli.Context) {
	prepare(c)

	if c.String(OPT_SKELETON) != "" {
		err := StoreSkeletonEC2RunConfigYaml(c.String(OPT_SKELETON))
		if err != nil {
			log.Printf("can not store Skeleton config yaml: %v\n", err)
		}

		os.Exit(0)
	}

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

	args := c.Args()
	if len(args) < 1 {
		log.Fatal("required ec2 run config file.")
	}

	cList := make([]EC2RunConfig, 0, len(args))
	for _, confPath := range args {
		configs := make([]EC2RunConfig, 0, 1)
		err := cstore.LoadFromYamlFile(confPath, &configs)
		if err != nil {
			log.Fatalf("failed load conf file: %v", err)
		}

		cList = append(cList, configs...)
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	specifiedName := c.String(OPT_SPECIFY_NAME)

	for _, conf := range cList {
		if specifiedName != "" && specifiedName != conf.Name {
			continue
		}

		tags := make([]*ec2.Tag, 0, len(conf.Tags))
		for _, t := range conf.Tags {
			ec2t := ec2.Tag{
				Key:   aws.String(t.Key),
				Value: aws.String(t.Value),
			}
			tags = append(tags, &ec2t)
		}

		launcher := conf.genLauncher()

		// overwrite launch parameter with command options
		if c.String(OPT_AMI_ID) != "" {
			launcher.AmiId = c.String(OPT_AMI_ID)
		}

		for i, l := range conf.Launches {
			// name replace check before launch instance
			// because name template fail, the instance is no Name tag instance.
			nr := &NameTagReplacement{
				Symbol:   c.String(OPT_SYMBOL),
				Sequence: strconv.Itoa(i + 1),
			}

			// instance type priority
			// command option > overwrite config > default config
			if c.String(OPT_I_TYPE) != "" {
				launcher.InstanceType = c.String(OPT_I_TYPE)
			} else {
				if l.OverWriteType != "" {
					launcher.InstanceType = l.OverWriteType
				} else {
					launcher.InstanceType = conf.Type
				}
			}

			replacedNameTag, err := nr.StringWithTemplate(l.NameTagTemplate)
			if err != nil {
				log.Fatalf("error during replacing name tag template: %v", err)
			}
			debug(replacedNameTag)

			res, err := launcher.Launch(cli, l.SubnetId, 1, c.Bool(OPT_DRYRUN))
			if err != nil {
				// TODO if dry run error then next.
				log.Fatalf("error during starting instance: %s", err.Error())
			}
			debug(res)

			nameTag := &ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(replacedNameTag),
			}

			for _, ins := range res.Instances {
				resources := make([]*string, 0, 3)
				resources = append(resources, ins.InstanceId)

				retrieveErrs := make([]error, 0, 3)
				for i = 0; i < 3; i++ {
					// Why sleep?
					// RunInstance result does not have BlockDeviceMappings
					// DescribeInstance that RunInstance same time too.
					// so sleep a second(or few second?)
					time.Sleep(500 * time.Millisecond)
					devMaps, err := myec2.GetBlockDeviceMappings(cli, *ins.InstanceId)

					if err != nil {
						retrieveErrs = append(retrieveErrs, err)
						continue
					}

					if len(devMaps) == 0 {
						retrieveErrs = append(retrieveErrs, fmt.Errorf("Not found DeviceMappings for: %s. it probably delaying device mapping.", ins.InstanceId))

						continue
					}

					for _, bdm := range devMaps {
						resources = append(resources, bdm.Ebs.VolumeId)
					}

					retrieveErrs = []error{}
					break
				}

				if len(retrieveErrs) > 0 {
					// currently, no handling failed tagging EBS
				}

				tagp := &ec2.CreateTagsInput{
					Resources: resources,
					// append returns new slice when over cap
					Tags: append(tags, nameTag),
				}

				_, err := cli.CreateTags(tagp)
				if err != nil {
					log.Println(fmt.Sprintf("failed tagging so skipped %s: %v\n", ins.InstanceId, err))
				}

				output := &EC2RunOutput{
					InstanceId: convertNilString(ins.InstanceId),
					Name:       replacedNameTag,
					PublicIp:   convertNilString(ins.PublicIpAddress),
					PrivateIp:  convertNilString(ins.PrivateIpAddress),
					Symbol:     nr.Symbol,
					Sequence:   nr.Sequence,
				}

				outputTemplate := DEFAULT_OUTPUT_TEMPLATE
				if l.OutputTemplate != "" {
					outputTemplate = l.OutputTemplate
				}

				oString, err := output.StringWithTemplate(outputTemplate)
				if err != nil {
					log.Println(fmt.Sprintf("%s failed replacing output template: %v", ins.InstanceId, err))
				}

				fmt.Println(oString)
			}
		}
	}
}

func doEc2Terminate(c *cli.Context) {
	prepare(c)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

	instanceId := c.String(OPT_INSTANCE_ID)
	var ids []*string
	if instanceId == "" {

		h, err := NewRnzooCStoreManager()
		if err != nil {
			log.Printf("can not load EC2: %s\n", err.Error())
		}

		fState := myec2.EC2_STATE_STOPPED
		if c.Bool(OPT_EC2_ANY_STATE) {
			fState = myec2.EC2_STATE_ANY
		}
		ids, err = h.ChooseEC2(region, fState, true)
		if err != nil {
			log.Fatalf("error during selecting: %s", err.Error())
			return
		}

	} else {
		ids = []*string{aws.String(instanceId)}
	}

	if len(ids) == 0 {
		log.Fatalln("there is no instance id.")
		return
	}

	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	if !c.Bool(OPT_WITHOUT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, ids...)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		for _, ins := range insts {
			name := "[no Name tag instance]"
			for _, t := range ins.Tags {
				if convertNilString(t.Key) == "Name" {
					name = convertNilString(t.Value)
					break
				}
			}

			fmt.Printf("%s\t%s\t%s\n", convertNilString(ins.InstanceId), name, convertNilString(ins.PrivateIpAddress))
		}

		ans, err := confirm("you really want to terminate above instances?", false)
		if !ans {
			log.Fatalln("canceled instance termination.")
			return
		}
	}

	dryrun := true
	if !c.Bool(OPT_DRYRUN) && c.Bool(OPT_EXECUTE) {
		dryrun = false
	}
	params := &ec2.TerminateInstancesInput{
		InstanceIds: ids,
		DryRun:      aws.Bool(dryrun),
	}

	resp, err := cli.TerminateInstances(params)
	if err != nil {
		log.Fatalf("error during terminate instance: %v", err)
		return
	}

	for _, status := range resp.TerminatingInstances {
		id := convertNilString(status.InstanceId)
		pState := convertNilString(status.PreviousState.Name)
		cState := convertNilString(status.CurrentState.Name)
		log.Printf("terminated %s: %s -> %s", id, pState, cState)
	}
}

func confirm(msg string, defaultAns bool) (bool, error) {

	if defaultAns {
		fmt.Printf("%s[YES/no]:", msg)
	} else {
		fmt.Printf("%s[yes/NO]:", msg)
	}

	reader := bufio.NewReader(os.Stdin)

	readAns, err := reader.ReadString('\n')
	if err != nil {
		return defaultAns, fmt.Errorf("input err:%s", err.Error())
	}

	inAns := strings.TrimRight(readAns, "\n")
	if inAns == "" {
		return defaultAns, nil
	}

	lAns := strings.ToLower(inAns)
	if defaultAns {
		return lAns == "no", nil
	} else {
		return lAns == "yes", nil
	}
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
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
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
			Usage: "this option was replaced. please use move-eip subcommand.",
		},
	},
}

var commandMoveEIP = cli.Command{
	Name:        "move-eip",
	Usage:       "reallocate EIP(allow reassociate) to other instance.",
	Description: "reallocate EIP(allow reassociate) to other instance.",
	Action:      doMoveEIP,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_CONFIRM,
			Usage: "without confirm target before action (default action is do confirming)",
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
			Name:  OPT_REGION + ", r",
			Usage: EC2LIST_REGION_USAGE,
		},
		cli.StringFlag{
			Name:  OPT_INSTANCE_ID,
			Usage: "specify instance id.",
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_RELEASE,
			Usage: "does not release disassociated the address.",
		},
		cli.BoolFlag{
			Name:  OPT_WITHOUT_CONFIRM,
			Usage: "without confirm target before action (default action is do confirming)",
		},
	},
}

func doMoveEIP(c *cli.Context) {
	prepare(c)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

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

	if !c.Bool(OPT_WITHOUT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, &instanceId)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		if len(insts) != 1 {
			log.Fatalln("the selected from instance was deleted? please retry.")
			return
		}

		name := "[no Name tag instance]"
		for _, t := range insts[0].Tags {
			if convertNilString(t.Key) == "Name" {
				name = convertNilString(t.Value)
				break
			}
		}

		eip := allocIds[0].PublicIP
		from := allocIds[0].Name
		to := name
		fmt.Printf("%s '%s' -> '%s'\n", eip, from, to)

		ans, err := confirm("move above EIP?", false)
		if !ans {
			log.Fatalln("canceled move EIP action.")
			return
		}
	}

	assocId, err := myec2.AssociateEIP(cli, allocIds[0].AllocationId, instanceId)
	if err != nil {
		log.Fatalf("error during moving EIP: %s", err.Error())
	}

	log.Printf("associated association_id:%s\tpublic_ip:%s\tinstance_id:%s", convertNilString(assocId), "EIP", instanceId)
}

// allocate new EIP and associate.
func doAttachEIP(c *cli.Context) {
	prepare(c)

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

	if c.Bool(OPT_MOVE) {
		log.Fatalln("this option was replaced. please use move-eip subcommand.")
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

	region, err := getRegion(c)
	if err != nil {
		log.Fatalln(err)
	}

	withoutRelease := c.Bool(OPT_WITHOUT_RELEASE)

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

	if !c.Bool(OPT_WITHOUT_CONFIRM) {
		insts, err := myec2.GetInstancesFromId(cli, &instanceId)
		if err != nil {
			log.Fatalln("failed retrieve instance info for confirm.")
			return
		}

		if len(insts) != 1 {
			log.Fatalln("the selected from instance was deleted? please retry.")
			return
		}

		name := "[no Name tag instance]"
		for _, t := range insts[0].Tags {
			if convertNilString(t.Key) == "Name" {
				name = convertNilString(t.Value)
				break
			}
		}

		fmt.Printf("%s\t%s\n", name, convertNilString(address.PublicIp))

		ans, err := confirm("you really want to detach above EIP?", false)
		if !ans {
			log.Fatalln("canceled detach EIP action.")
			return
		}
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

// cloudwatch
var commandGetBilling = cli.Command{
	Name:        "billing-price",
	ShortName:   "price",
	Usage:       "show billing price EstimatedCharges (CAUTION: NOT real time)",
	Description: `the billing price is get from CloudWatch AWS/Billing.`,
	Action:      doShowBilling,
	Flags:       []cli.Flag{},
}

func doShowBilling(c *cli.Context) {
	prepare(c)

	b, err := GetBillingEstimatedCharges()
	if err != nil {
		log.Fatalf("failed get billing price: %v", err)
	}

	log.Printf("%s %.2f USD\n", b.Label, b.Price)
}

func NewCStoreManager() (*cstore.Manager, error) {
	dirPath, err := GetRnzooDir()
	if err != nil {
		return nil, err
	}
	return cstore.NewManager("rnzoo", dirPath)
}

func NewRnzooCStoreManager() (*myec2.EC2Handler, error) {
	m, err := NewCStoreManager()
	if err != nil {
		return nil, err
	}

	return myec2.NewEC2Handler(m), nil
}
