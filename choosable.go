package main

import (
	"fmt"

	"github.com/reiki4040/peco"
)

func ChooseEC2(region, msg string) ([]string, error) {
	instances, err := GetInstances(region)
	if err != nil {
		return nil, err
	}

	choiceSlice := make([]peco.Choosable, 0, len(instances))
	for _, i := range instances {
		var nameTag string
		for _, tag := range i.Tags {
			if convertNilString(tag.Key) == "Name" {
				nameTag = convertNilString(tag.Value)
				break
			}
		}

		id := convertNilString(i.InstanceId)
		choice := fmt.Sprintf("instance_id:%s\tname:%s\tstate:%s\tpublic_ip:%s\tprivate_ip:%s\n",
			id,
			nameTag,
			convertNilString(i.State.Name),
			convertNilString(i.PublicIpAddress),
			convertNilString(i.PrivateIpAddress))

		c := &peco.Choice{
			C: choice,
			V: id,
		}

		choiceSlice = append(choiceSlice, c)
	}

	chosen, err := peco.Choose("EC2 instance", "please select EC instance", choiceSlice)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(chosen))
	for _, c := range chosen {
		ids = append(ids, c.Value())
	}

	return ids, nil
}
