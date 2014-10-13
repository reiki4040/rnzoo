package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	flag "github.com/dotcloud/docker/pkg/mflag"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

const (
	Version = "0.1.0"
	Usage   = `ec2list
  ## Description
  
  show your ec2 infomations with LTSV format.
  this command make cache file that ec2 info. (default ~/.rnzoo/instance.cache.REGION)
  second time, you can get ec2 info without access to AWS.
  
  ## Usage
  
  show your ec2 info at ap-northeast-1
  
      ec2list -r ap-northeast-1
  
      instance_id:i-1111111       name:Name tags Value1   state:stopped   public_ip:X.X.X.X       private_ip:Y.Y.Y.Y      instance_type:t2.micro
      instance_id:i-22222222      name:Name tags Value2   state:running   public_ip:X.X.X.x       private_ip:Y.Y.Y.y      instance_type:m3.large
      ...
  
  if you updated ec2(create new instance, stop, start and etc...), need to update cache with -f/--force option.
  
      ec2list -r ap-northeast-1 -f
  
  you can set default region by AWS_REGION environment variable.
  
      export AWS_REGION=ap-northeast-1`

	CACHE_PATH_PREFIX = "/Users/reiki/.rnzoo/instances.cache."
)

var (
	force_reload bool
	regionName   string
	show_version bool
	show_usage   bool
)

func parseflg() {
	flag.BoolVar(&show_version, []string{"v", "-version"}, false, "show version.")
	flag.BoolVar(&show_usage, []string{"h", "-help"}, false, "show this usage.")
	flag.BoolVar(&force_reload, []string{"f", "-force"}, false, "reload ec2 (force connect to AWS)")
	flag.StringVar(&regionName, []string{"r", "-region"}, "", "specify region")
	flag.Parse()
}

func version() {
	log.Printf("%s\n", Version)
}

func usage() {
	log.Printf("%s\n", Usage)
}

func main() {
	parseflg()

	if show_usage {
		usage()
		os.Exit(0)
	}

	if show_version {
		version()
		os.Exit(0)
	}

	if regionName == "" {
		regionName = os.Getenv("AWS_REGION")
	}

	region, err := GetRegion(regionName)
	if err != nil {
		log.Fatalf("failed region: %s", err.Error())
	}

	ec2list(force_reload, region)
}

func GetRegion(region string) (*aws.Region, error) {
	switch strings.ToLower(region) {
	case "ap-northeast-1":
		return &aws.APNortheast, nil
		// TODO other region
	default:
		return nil, errors.New(fmt.Sprintf("invalid region name: %s", region))
	}
}

func ec2list(reload bool, region *aws.Region) {
	var instances []ec2.Instance
	cachePath := CACHE_PATH_PREFIX + region.Name
	if _, err := os.Stat(cachePath); os.IsNotExist(err) || reload {
		auth, err := aws.EnvAuth()
		if err != nil {
			log.Fatalf("failed auth: %s\n", err.Error())
		}
		instances, err = GetInstances(auth, region)
		if err != nil {
			log.Fatalf("failed get instance: %s", err.Error())
		}

		err = StoreCache(instances, cachePath)
		if err != nil {
			log.Printf("failed store cache: %s", err.Error())
		}
	} else {
		instances, err = LoadCache(cachePath)
		if err != nil {
			// TODO retrieve from aws
			log.Fatalf("failed get instance from cache: %s", err.Error())
		}
	}

	for _, i := range instances {
		showLtsv(i)
	}
}

type Instances struct {
	Instances []ec2.Instance `xml:"Instance"`
}

func StoreCache(instances []ec2.Instance, cachePath string) error {
	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	w := bufio.NewWriter(cacheFile)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	toXml := Instances{Instances: instances}
	if err := enc.Encode(toXml); err != nil {
		return err
	}

	return nil
}

func LoadCache(cachePath string) ([]ec2.Instance, error) {
	cacheFile, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer cacheFile.Close()

	r := bufio.NewReader(cacheFile)
	dec := xml.NewDecoder(r)
	instances := Instances{}
	err = dec.Decode(&instances)
	if err != nil {
		return nil, err
	}

	return instances.Instances, nil
}

func GetInstances(auth aws.Auth, region *aws.Region) ([]ec2.Instance, error) {
	ec2conn := ec2.New(auth, aws.APNortheast)

	resp, err := ec2conn.DescribeInstances(nil, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return nil, errors.New("there is no instance.")
	}

	instances := make([]ec2.Instance, 0)
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, i)
		}
	}

	return instances, nil
}

func showLtsv(i ec2.Instance) {
	var nameTag string
	for _, tag := range i.Tags {
		if tag.Key == "Name" {
			nameTag = tag.Value
			break
		}
	}

	fmt.Printf("instance_id:%s\tname:%s\tstate:%s\tpublic_ip:%s\tprivate_ip:%s\n",
		i.InstanceId,
		nameTag,
		i.State.Name,
		i.IPAddress,
		i.PrivateIPAddress)
}
