package ec2

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/reiki4040/cstore"
	"github.com/reiki4040/peco"
)

const (
	EC2_LIST_CACHE_PREFIX = "aws.instances.cache."

	EC2_STATE_ANY     = ""
	EC2_STATE_RUNNING = "running"
	EC2_STATE_STOPPED = "stopped"
)

func MakeEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return ec2.NewFromConfig(cfg), nil
}

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
	Instances []types.Instance `json:"ec2_instances"`
}

func NewEC2Handler(m *cstore.Manager) *EC2Handler {
	return &EC2Handler{
		Manager: m,
	}
}

type EC2Handler struct {
	Manager *cstore.Manager
}

func (h *EC2Handler) ChooseEC2(region, state string, reload bool) ([]string, error) {
	ec2list, err := h.LoadChoosableEC2List(region, state, reload)
	if err != nil {
		return nil, err
	}

	choices := ConvertChoosableList(ec2list)

	chosens, err := peco.Choose("EC2", "select instances", "", choices)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(chosens))
	for _, c := range chosens {
		if ec2, ok := c.(*ChoosableEC2); ok {
			ids = append(ids, ec2.InstanceId)
		}
	}

	return ids, nil
}

func (r *EC2Handler) GetCacheStore(region string) (*cstore.CStore, error) {
	cacheFileName := EC2_LIST_CACHE_PREFIX + region + ".json"
	return r.Manager.New(cacheFileName, cstore.JSON)
}

func (r *EC2Handler) LoadChoosableEC2List(region, state string, reload bool) ([]*ChoosableEC2, error) {
	var instances []types.Instance
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

func ConvertChoosableEC2List(instances []types.Instance, state string) []*ChoosableEC2 {
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

func convertChoosable(ins types.Instance) *ChoosableEC2 {

	var nameTag string
	for _, tag := range ins.Tags {
		if convertNilString(tag.Key) == "Name" {
			nameTag = convertNilString(tag.Value)
			break
		}
	}

	ipv6 := ""
	for _, ni := range ins.NetworkInterfaces {
		for _, v6addr := range ni.Ipv6Addresses {
			if v6 := convertNilString(v6addr.Ipv6Address); v6 != "" {
				ipv6 = v6
				break
			}
		}
	}
	c := &ChoosableEC2{
		InstanceId:   convertNilString(ins.InstanceId),
		Name:         nameTag,
		Status:       convertNilString((*string)(&ins.State.Name)),
		InstanceType: convertNilString((*string)(&ins.InstanceType)),
		PublicIP:     convertNilString(ins.PublicIpAddress),
		PrivateIP:    convertNilString(ins.PrivateIpAddress),
		IPv6:         ipv6,
	}

	return c
}

func GetInstances(region string) ([]types.Instance, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	cli := ec2.NewFromConfig(cfg)
	resp, err := cli.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return []types.Instance{}, nil
	}

	instances := make([]types.Instance, 0)
	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, i)
		}
	}

	return instances, nil
}

func GetInstancesFromId(ctx context.Context, cli *ec2.Client, ids ...string) ([]types.Instance, error) {
	param := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}

	resp, err := cli.DescribeInstances(ctx, param)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return []types.Instance{}, nil
	}

	instances := make([]types.Instance, 0)
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

func ChooseEIP(ctx context.Context, region string) ([]*ChoosableEIP, error) {
	EIPs, err := LoadEIPList(ctx, region)
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

func LoadEIPList(ctx context.Context, region string) ([]*ChoosableEIP, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	cli := ec2.NewFromConfig(cfg)
	resp, err := cli.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
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

func AssociateEIP(ctx context.Context, cli *ec2.Client, eipAllocId, instanceId string) (*string, error) {
	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(eipAllocId),
		AllowReassociation: aws.Bool(true),
		InstanceId:         aws.String(instanceId),
	}
	resp, err := cli.AssociateAddress(ctx, params)

	return resp.AssociationId, err
}

func AllocateEIP(ctx context.Context, cli *ec2.Client) (*string, *string, error) {
	params := &ec2.AllocateAddressInput{
		Domain: types.DomainTypeVpc,
	}
	resp, err := cli.AllocateAddress(ctx, params)
	return resp.AllocationId, resp.PublicIp, err
}

func DisassociateEIP(ctx context.Context, cli *ec2.Client, allocId string) error {
	params := &ec2.DisassociateAddressInput{
		AssociationId: aws.String(allocId),
	}

	// resp is empty struct
	_, err := cli.DisassociateAddress(ctx, params)

	return err
}

func ReleaseEIP(ctx context.Context, cli *ec2.Client, allocId string) error {
	params := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(allocId),
	}

	// resp is empty struct
	_, err := cli.ReleaseAddress(ctx, params)
	return err
}

func GetEIPFromInstance(ctx context.Context, cli *ec2.Client, instanceId string) (*types.Address, error) {
	params := &ec2.DescribeAddressesInput{
		Filters: []types.Filter{
			types.Filter{
				Name: aws.String("instance-id"),
				Values: []string{
					instanceId,
				},
			},
		},
	}
	resp, err := cli.DescribeAddresses(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Addresses) != 1 {
		return nil, errors.New("this instance has not EIP.")
	}

	address := resp.Addresses[0]
	return &address, nil
}

func GetNotAssociateEIP(ctx context.Context, cli *ec2.Client) (*types.Address, error) {
	params := &ec2.DescribeAddressesInput{}

	resp, err := cli.DescribeAddresses(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(resp.Addresses) >= 1 {
		for _, address := range resp.Addresses {
			if address.InstanceId == nil {
				return &address, nil
			}
		}
	}

	return nil, nil
}

type Launcher struct {
	AmiId              string
	InstanceType       string
	KeyName            string
	SecurityGroupIds   []string
	PublicIpEnabled    bool
	Ipv6Enabled        bool
	IamRoleName        *string
	EbsDevices         []Ebs
	EbsOptimized       bool
	PlacementGroupName string
	UserData           string
}

// why encrypted use *bool?
// for modify root device volume size. cannot specify encrypted root device
type Ebs struct {
	DeviceName          string
	DeleteOnTermination bool
	Encrypted           *bool
	SizeGB              int64
	VolumeType          string
}

func (d *Launcher) Launch(ctx context.Context, cli *ec2.Client, subnetId string, count int, dryrun bool) (*ec2.RunInstancesOutput, error) {
	var ebsMappings []types.BlockDeviceMapping
	if len(d.EbsDevices) > 0 {
		ebsMappings = make([]types.BlockDeviceMapping, 0, len(d.EbsDevices))
		for _, ebs := range d.EbsDevices {
			m := types.BlockDeviceMapping{
				DeviceName: aws.String(ebs.DeviceName),
				Ebs: &types.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(ebs.DeleteOnTermination),
					Encrypted:           ebs.Encrypted,
					//Iops:                aws.Int64(100),
					//SnapshotId:          aws.String("String"),
					VolumeSize: aws.Int32(int32(ebs.SizeGB)),
					VolumeType: types.VolumeType(ebs.VolumeType),
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

	var placement *types.Placement
	if d.PlacementGroupName != "" {
		placement = &types.Placement{
			GroupName: aws.String(d.PlacementGroupName),
		}
	}

	var userData string
	if d.UserData != "" {
		userData = base64.StdEncoding.EncodeToString([]byte(d.UserData))
	}

	params := &ec2.RunInstancesInput{
		ImageId:             aws.String(d.AmiId),
		MaxCount:            aws.Int32(int32(count)),
		MinCount:            aws.Int32(int32(count)),
		BlockDeviceMappings: ebsMappings,
		//AdditionalInfo: aws.String("String"),
		//ClientToken:           aws.String("String"),
		//DisableApiTermination: aws.Bool(true),
		DryRun:       aws.Bool(dryrun),
		EbsOptimized: aws.Bool(d.EbsOptimized),
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			//Arn: aws.String("arn:aws:iam::<aws_id>:instance-profile/sample_iamrole"),
			Name: d.IamRoleName,
		},
		InstanceType: types.InstanceType(d.InstanceType),
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
		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			types.InstanceNetworkInterfaceSpecification{
				AssociatePublicIpAddress: aws.Bool(d.PublicIpEnabled),
				DeviceIndex:              aws.Int32(0),
				SubnetId:                 aws.String(subnetId),
				Groups:                   d.SecurityGroupIds,
				Ipv6AddressCount:         aws.Int32(int32(*ipv6count)),
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
		Placement: placement,
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
		UserData: aws.String(userData),
	}

	return cli.RunInstances(ctx, params)
}

func GetBlockDeviceMappings(ctx context.Context, cli *ec2.Client, instanceId string) ([]types.InstanceBlockDeviceMapping, error) {
	descIns, err := GetInstancesFromId(ctx, cli, instanceId)
	if err != nil {
		return nil, err
	}

	if len(descIns) > 0 {
		return descIns[0].BlockDeviceMappings, nil
	} else {
		return nil, nil
	}
}

func convertNilString(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}
