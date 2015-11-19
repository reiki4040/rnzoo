package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"
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

	ec2list(regionName, isReload)
}

func GetRnzooDir() string {
	rnzooDir := os.Getenv(ENV_HOME) + string(os.PathSeparator) + RNZOO_DIR_NAME
	return rnzooDir
}

func GetEC2ListCachePath(region string) string {
	rnzooDir := GetRnzooDir()
	return rnzooDir + string(os.PathSeparator) + "aws.instances.cache." + region
}

func CreateRnzooDir() error {
	rnzooDir := GetRnzooDir()

	if _, err := os.Stat(rnzooDir); os.IsNotExist(err) {
		err = os.Mkdir(rnzooDir, 0700)
		if err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}

	return nil
}

func ec2list(region string, reload bool) {
	var instances []*ec2.Instance
	cachePath := GetEC2ListCachePath(region)

	if _, err := os.Stat(cachePath); os.IsNotExist(err) || reload {
		var err error
		instances, err = GetInstances(region)
		if err != nil {
			log.Fatalf("failed get instance: %s", err.Error())
		}

		if err != nil {
			log.Fatalf("failed get instance: %s", err.Error())
		}

		err = StoreCache(instances, cachePath)
		if err != nil {
			// only warn message
			log.Printf("warn: failed store ec2 list cache: %s\n", err.Error())
		}
	} else {
		var err error
		instances, err = LoadCache(cachePath)
		if err != nil {
			// only warn message
			log.Printf("warn: failed load ec2 list cache: %s, so try load from AWS.\n", err.Error())

			instances, err = GetInstances(region)
			if err != nil {
				log.Fatalf("failed get instance: %s", err.Error())
			}
		}
	}

	for _, i := range instances {
		showLtsv(i)
	}
}

type Instances struct {
	Instances []*ec2.Instance `json:"ec2_instances"`
}

func StoreCache(instances []*ec2.Instance, cachePath string) error {
	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	w := bufio.NewWriter(cacheFile)
	enc := json.NewEncoder(w)
	//enc.Indent("", "  ")
	toJson := Instances{Instances: instances}
	if err := enc.Encode(toJson); err != nil {
		return err
	}

	return nil
}

func LoadCache(cachePath string) ([]*ec2.Instance, error) {
	cacheFile, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer cacheFile.Close()

	r := bufio.NewReader(cacheFile)
	dec := json.NewDecoder(r)
	instances := Instances{}
	err = dec.Decode(&instances)
	if err != nil {
		return nil, err
	}

	return instances.Instances, nil
}

func GetInstances(region string) ([]*ec2.Instance, error) {
	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	resp, err := cli.DescribeInstances(nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return []*ec2.Instance{}, nil
	}

	instances := make([]*ec2.Instance, 0)
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, i)
		}
	}

	return instances, nil
}

func showLtsv(i *ec2.Instance) {

	var nameTag string
	for _, tag := range i.Tags {
		if convertNilString(tag.Key) == "Name" {
			nameTag = convertNilString(tag.Value)
			break
		}
	}

	//ins := *i
	fmt.Printf("instance_id:%s\tname:%s\tstate:%s\tpublic_ip:%s\tprivate_ip:%s\n",
		convertNilString(i.InstanceId),
		nameTag,
		convertNilString(i.State.Name),
		convertNilString(i.PublicIpAddress),
		convertNilString(i.PrivateIpAddress))
}
