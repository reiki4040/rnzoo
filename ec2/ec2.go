package ec2

import (
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/reiki4040/cstore"
	"github.com/reiki4040/peco"
)

const (
	EC2_LIST_CACHE_PREFIX = "aws.instances.cache."
)

func ConvertChoosableList(ec2List []*ChoosableEC2) []peco.Choosable {
	choices := make([]peco.Choosable, 0, len(ec2List))
	for _, c := range ec2List {
		choices = append(choices, c)
	}
	return choices
}

type ChoosableEC2 struct {
	InstanceId   string
	Name         string
	Status       string
	InstanceType string
	PublicIP     string
	PrivateIP    string
}

func (e *ChoosableEC2) Choice() string {
	w := new(tabwriter.Writer)
	var b bytes.Buffer
	w.Init(&b, 18, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s", e.InstanceId, e.Name, e.Status, e.InstanceType, e.PublicIP, e.PrivateIP)
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

func (h *EC2Handler) ChooseEC2(region string, reload bool) ([]*string, error) {
	ec2list, err := h.LoadChoosableEC2List(region, reload)
	if err != nil {
		return nil, err
	}

	choices := ConvertChoosableList(ec2list)

	chosens, err := peco.Choose("EC2", "select instances", "", choices)
	if err != nil {
		return nil, err
	}

	ids := make([]*string, 0, len(chosens))
	for _, c := range chosens {
		if ec2, ok := c.(*ChoosableEC2); ok {
			ids = append(ids, aws.String(ec2.InstanceId))
		}
	}

	return ids, nil
}

func (r *EC2Handler) GetCacheStore(region string) (*cstore.CStore, error) {
	cacheFileName := EC2_LIST_CACHE_PREFIX + region + ".json"
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
		InstanceId:   convertNilString(ins.InstanceId),
		Name:         nameTag,
		Status:       convertNilString(ins.State.Name),
		InstanceType: convertNilString(ins.InstanceType),
		PublicIP:     convertNilString(ins.PublicIpAddress),
		PrivateIP:    convertNilString(ins.PrivateIpAddress),
	}

	return c
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

func convertNilString(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}
