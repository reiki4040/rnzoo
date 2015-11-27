package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/codegangsta/cli"

	"github.com/reiki4040/cstore"
	"github.com/reiki4040/peco"
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

func NewRnzooCStoreManager() (*EC2Handler, error) {
	dirPath := GetRnzooDir()
	m, err := cstore.NewManager("rnzoo", dirPath)
	if err != nil {
		return nil, err
	}

	return NewEC2Handler(m), nil
}

func ConvertChoosableList(ec2List []*ChoosableEC2) []peco.Choosable {
	choices := make([]peco.Choosable, 0, len(ec2List))
	for _, c := range ec2List {
		choices = append(choices, c)
	}
	return choices
}

type ChoosableEC2 struct {
	InstanceId string
	Name       string
	Status     string
	PublicIP   string
	PrivateIP  string
}

func (e *ChoosableEC2) Choice() string {
	w := new(tabwriter.Writer)
	var b bytes.Buffer
	w.Init(&b, 18, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s", e.InstanceId, e.Name, e.Status, e.PublicIP, e.PrivateIP)
	w.Flush()
	return string(b.Bytes())
}

func (e *ChoosableEC2) Value() string {
	return e.InstanceId
}

type ChoosableEC2s []*ChoosableEC2

func (e ChoosableEC2s) Len() int {
	return len(e)
}

func (e ChoosableEC2s) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e ChoosableEC2s) Less(i, j int) bool {
	return e[i].Name < e[j].Name
}

type Instances struct {
	Instances []*ec2.Instance `json:"ec2_instances"`
}

func NewEC2Handler(m *cstore.Manager) *EC2Handler {
	return &EC2Handler{
		Manager: m,
	}
}

type EC2Handler struct {
	Manager *cstore.Manager
}

func (r *EC2Handler) GetCacheStore(region string) (*cstore.CStore, error) {
	cacheFileName := RNZOO_EC2_LIST_CACHE_PREFIX + region + ".json"
	return r.Manager.New(cacheFileName, cstore.JSON)
}

func (r *EC2Handler) LoadChoosableEC2List(region string, reload bool) ([]*ChoosableEC2, error) {
	var instances []*ec2.Instance
	cacheStore, _ := r.GetCacheStore(region)

	is := Instances{}
	if cErr := cacheStore.GetWithoutValidate(&is); cErr != nil || reload {
		var err error
		instances, err = GetInstances(region)
		if err != nil {
			awsErr := fmt.Errorf("failed get instance: %s", err.Error())
			return nil, awsErr
		}

		is = Instances{Instances: instances}
		if cacheStore != nil {
			err := cacheStore.SaveWithoutValidate(&is)
			if err != nil {
				// only warn message
				fmt.Printf("warn: failed store ec2 list cache: %s\n", err.Error())
			}
		}
	}

	choices := ConvertChoosableEC2List(is.Instances)
	if len(choices) == 0 {
		err := fmt.Errorf("there is no running instance.")
		return nil, err
	}

	return choices, nil
}

func ConvertChoosableEC2List(instances []*ec2.Instance) []*ChoosableEC2 {
	choosableEC2List := make([]*ChoosableEC2, 0, len(instances))
	for _, i := range instances {
		e := convertChoosable(i)
		if e != nil {
			choosableEC2List = append(choosableEC2List, e)
		}
	}

	sort.Sort(ChoosableEC2s(choosableEC2List))
	return choosableEC2List
}

func convertChoosable(i *ec2.Instance) *ChoosableEC2 {

	var nameTag string
	for _, tag := range i.Tags {
		if convertNilString(tag.Key) == "Name" {
			nameTag = convertNilString(tag.Value)
			break
		}
	}

	ins := *i
	c := &ChoosableEC2{
		InstanceId: convertNilString(ins.InstanceId),
		Name:       nameTag,
		Status:     convertNilString(ins.State.Name),
		PublicIP:   convertNilString(ins.PublicIpAddress),
		PrivateIP:  convertNilString(ins.PrivateIpAddress),
	}

	return c
}

func GetEC2ListCachePath(region string) string {
	rnzooDir := GetRnzooDir()
	return rnzooDir + string(os.PathSeparator) + "aws.instances.cache." + region
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
