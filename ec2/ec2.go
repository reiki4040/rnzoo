package ec2

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/reiki4040/cstore"
	"github.com/reiki4040/peco"
)

const (
	EC2_LIST_CACHE_PREFIX = "aws.instances.cache."

	EC2_STATE_ANY     = ""
	EC2_STATE_RUNNING = "running"
	EC2_STATE_STOPPED = "stopped"
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
	IPv6         string
}

func (e *ChoosableEC2) Choice() string {
	w := new(tabwriter.Writer)
	var b bytes.Buffer
	w.Init(&b, 18, 0, 4, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s", e.InstanceId, e.Name, e.Status, e.InstanceType, e.PublicIP, e.PrivateIP, e.IPv6)
	w.Flush()
	return string(b.Bytes())
}

func (e *ChoosableEC2) Value() string {
	return e.InstanceId
}

func (e *ChoosableEC2) String() string {
	items := []string{e.InstanceId, e.Name, e.Status, e.InstanceType, e.PublicIP, e.PrivateIP, e.IPv6}
	return strings.Join(items, "\t")
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

func (h *EC2Handler) ChooseEC2(region, state string, reload bool) ([]*string, error) {
	ec2list, err := h.LoadChoosableEC2List(region, state, reload)
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

func (r *EC2Handler) LoadChoosableEC2List(region, state string, reload bool) ([]*ChoosableEC2, error) {
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

	choices := ConvertChoosableEC2List(is.Instances, state)
	if len(choices) == 0 {
		err := fmt.Errorf("there is no instance.")
		return nil, err
	}

	return choices, nil
}

func ConvertChoosableEC2List(instances []*ec2.Instance, state string) []*ChoosableEC2 {
	choosableEC2List := make([]*ChoosableEC2, 0, len(instances))
	for _, i := range instances {
		e := convertChoosable(i)
		if e != nil {
			if state != EC2_STATE_ANY {
				if e.Status != state {
					continue
				}
			}

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

	ipv6 := ""
	for _, ni := range i.NetworkInterfaces {
		for _, v6addr := range ni.Ipv6Addresses {
			if v6 := convertNilString(v6addr.Ipv6Address); v6 != "" {
				ipv6 = v6
				break
			}
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
		IPv6:         ipv6,
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

func GetInstancesFromId(cli *ec2.EC2, ids ...*string) ([]*ec2.Instance, error) {
	param := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}

	resp, err := cli.DescribeInstances(param)
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

type ChoosableEIP struct {
	AllocationId string
	AssociateId  string
	PublicIP     string
	InstanceId   string
	Name         string
}

func (c *ChoosableEIP) Choice() string {
	return fmt.Sprintf("%s %s", c.PublicIP, c.Name)
}

func (c *ChoosableEIP) Value() string {
	return c.AllocationId
}

func ChooseEIP(region string) ([]*ChoosableEIP, error) {
	EIPs, err := LoadEIPList(region)
	if err != nil {
		return nil, err
	}

	choices := ConvertChoosableEIPList(EIPs)

	chosens, err := peco.Choose("EIP", "select EIP", "", choices)
	if err != nil {
		return nil, err
	}

	ids := make([]*ChoosableEIP, 0, len(chosens))
	for _, c := range chosens {
		if eip, ok := c.(*ChoosableEIP); ok {
			ids = append(ids, eip)
		}
	}

	return ids, nil
}

func ConvertChoosableEIPList(eipList []*ChoosableEIP) []peco.Choosable {
	choices := make([]peco.Choosable, 0, len(eipList))
	for _, e := range eipList {
		choices = append(choices, e)
	}
	return choices
}

func LoadEIPList(region string) ([]*ChoosableEIP, error) {
	cli := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})
	resp, err := cli.DescribeAddresses(nil)
	if err != nil {
		return nil, err
	}

	instances, err := GetInstances(region)
	if err != nil {
		return nil, err
	}

	iMap := make(map[string]string, 0)
	for _, i := range instances {
		cEC2 := convertChoosable(i)
		if cEC2.InstanceId != "" {
			iMap[cEC2.InstanceId] = cEC2.Name
		}
	}

	cEIPs := make([]*ChoosableEIP, 0)
	for _, addr := range resp.Addresses {
		name, _ := iMap[convertNilString(addr.InstanceId)]
		e := &ChoosableEIP{
			AllocationId: convertNilString(addr.AllocationId),
			AssociateId:  convertNilString(addr.AssociationId),
			PublicIP:     convertNilString(addr.PublicIp),
			InstanceId:   convertNilString(addr.InstanceId),
			Name:         name,
		}

		cEIPs = append(cEIPs, e)
	}

	return cEIPs, nil
}

func AssociateEIP(cli *ec2.EC2, eipAllocId, instanceId string) (*string, error) {
	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(eipAllocId),
		AllowReassociation: aws.Bool(true),
		InstanceId:         aws.String(instanceId),
	}
	resp, err := cli.AssociateAddress(params)

	return resp.AssociationId, err
}

func AllocateEIP(cli *ec2.EC2) (*string, *string, error) {
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	}
	resp, err := cli.AllocateAddress(params)
	return resp.AllocationId, resp.PublicIp, err
}

func DisassociateEIP(cli *ec2.EC2, allocId string) error {
	params := &ec2.DisassociateAddressInput{
		AssociationId: aws.String(allocId),
	}

	// resp is empty struct
	_, err := cli.DisassociateAddress(params)

	return err
}

func ReleaseEIP(cli *ec2.EC2, allocId string) error {
	params := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(allocId),
	}

	// resp is empty struct
	_, err := cli.ReleaseAddress(params)
	return err
}

func GetEIPFromInstance(cli *ec2.EC2, instanceId string) (*ec2.Address, error) {
	params := &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("instance-id"),
				Values: []*string{
					aws.String(instanceId),
				},
			},
		},
	}
	resp, err := cli.DescribeAddresses(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Addresses) != 1 {
		return nil, errors.New("this instance has not EIP.")
	}

	address := resp.Addresses[0]
	return address, nil
}

func GetNotAssociateEIP(cli *ec2.EC2) (*ec2.Address, error) {
	params := &ec2.DescribeAddressesInput{}

	resp, err := cli.DescribeAddresses(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Addresses) >= 1 {
		for _, address := range resp.Addresses {
			if address.InstanceId == nil {
				return address, nil
			}
		}
	}

	return nil, nil
}

type Launcher struct {
	AmiId            string
	InstanceType     string
	KeyName          string
	SecurityGroupIds []*string
	PublicIpEnabled  bool
	Ipv6Enabled      bool
	IamRoleName      *string
	EbsDevices       []Ebs
	EbsOptimized     bool
}

type Ebs struct {
	DeviceName          string
	DeleteOnTermination bool
	Encrypted           bool
	SizeGB              int64
	VolumeType          string
}

func (d *Launcher) Launch(cli *ec2.EC2, subnetId string, count int, dryrun bool) (*ec2.Reservation, error) {
	var ebsMappings []*ec2.BlockDeviceMapping
	if len(d.EbsDevices) > 0 {
		ebsMappings = make([]*ec2.BlockDeviceMapping, 0, len(d.EbsDevices))
		for _, ebs := range d.EbsDevices {
			m := &ec2.BlockDeviceMapping{
				DeviceName: aws.String(ebs.DeviceName),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(ebs.DeleteOnTermination),
					Encrypted:           aws.Bool(ebs.Encrypted),
					//Iops:                aws.Int64(100),
					//SnapshotId:          aws.String("String"),
					VolumeSize: aws.Int64(ebs.SizeGB),
					VolumeType: aws.String(ebs.VolumeType),
				},
				//NoDevice:    aws.String("String"),
				//VirtualName: aws.String("String"),
			}

			ebsMappings = append(ebsMappings, m)
		}
	}

	var keyName *string
	if d.KeyName != "" {
		keyName = &d.KeyName
	}

	var ipv6count *int64
	if d.Ipv6Enabled {
		ipv6count = aws.Int64(1)
	}

	params := &ec2.RunInstancesInput{
		ImageId:             aws.String(d.AmiId),
		MaxCount:            aws.Int64(int64(count)),
		MinCount:            aws.Int64(int64(count)),
		BlockDeviceMappings: ebsMappings,
		//AdditionalInfo: aws.String("String"),
		//ClientToken:           aws.String("String"),
		//DisableApiTermination: aws.Bool(true),
		DryRun:       aws.Bool(dryrun),
		EbsOptimized: aws.Bool(d.EbsOptimized),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			//Arn: aws.String("arn:aws:iam::694958806517:instance-profile/sample_iamrole"),
			Name: d.IamRoleName,
		},
		InstanceType: aws.String(d.InstanceType),
		//Ipv6AddressCount: aws.Int64(1),
		//Ipv6Addresses: []*ec2.InstanceIpv6Address{
		//	{ // Required
		//		Ipv6Address: aws.String("String"),
		//	},
		//	// More values...
		//},
		//KernelId: aws.String("String"),
		KeyName: keyName,
		//Monitoring: &ec2.RunInstancesMonitoringEnabled{
		//	Enabled: aws.Bool(true), // Required
		//},
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			&ec2.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress: aws.Bool(d.PublicIpEnabled),
				DeviceIndex:              aws.Int64(0),
				SubnetId:                 aws.String(subnetId),
				Groups:                   d.SecurityGroupIds,
				Ipv6AddressCount:         ipv6count,
			},
		},
		//	{ // Required
		//		AssociatePublicIpAddress: aws.Bool(false),
		//		DeleteOnTermination:      aws.Bool(true),
		//		Description:              aws.String("String"),
		//		DeviceIndex:              aws.Int64(1),
		//		Groups: []*string{
		//			aws.String("String"), // Required
		//			// More values...
		//		},
		//		// SDK compile error IPv6... unknown field...
		//		//Ipv6AddressCount: aws.Int64(1),
		//		//Ipv6Addresses: []*ec2.InstanceIpv6Address{
		//		//	{ // Required
		//		//		Ipv6Address: aws.String("String"),
		//		//	},
		//		//	// More values...
		//		//},
		//		//NetworkInterfaceId: aws.String("String"),
		//		//PrivateIpAddress:   aws.String("String"),
		//		//PrivateIpAddresses: []*ec2.PrivateIpAddressSpecification{
		//		//	{ // Required
		//		//		PrivateIpAddress: aws.String("String"), // Required
		//		//		Primary:          aws.Bool(true),
		//		//	},
		//		//	// More values...
		//		//},
		//		//SecondaryPrivateIpAddressCount: aws.Int64(1),
		//		SubnetId: aws.String(subnetId),
		//	},
		//	// More values...
		//},
		//Placement: &ec2.Placement{
		//	Affinity:         aws.String("String"),
		//	AvailabilityZone: aws.String("String"),
		//	GroupName:        aws.String("String"),
		//	HostId:           aws.String("String"),
		//	Tenancy:          aws.String("Tenancy"),
		//},
		//PrivateIpAddress: aws.String("String"),
		//RamdiskId:        aws.String("String"),
		//SecurityGroupIds: p.SecurityGroupIds,
		//SubnetId:         aws.String(p.SubnetId),
		//UserData: aws.String("String"),
	}

	return cli.RunInstances(params)

}

func convertNilString(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}
