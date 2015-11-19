package main

import (
	"fmt"

	"github.com/reiki4040/peco"
)

type Choosable interface {
	Choice() string
	Value() string
}

type Choice struct {
	C string
	V string
}

func (c *Choice) Choice() string {
	return c.C
}

func (c *Choice) Value() string {
	return c.V
}

func Choose(itemName, message string, choices []Choosable) ([]Choosable, error) {
	if len(choices) == 0 {
		err := fmt.Errorf("there is no %s.", itemName)
		return nil, err
	}

	pecoChoices := make([]peco.Choosable, 0, len(choices))
	for _, c := range choices {
		pecoChoices = append(pecoChoices, c)
	}

	pecoOpt := &peco.PecoOptions{
		OptPrompt: fmt.Sprintf("%s >", message),
	}

	result, err := peco.PecolibWithOptions(pecoChoices, pecoOpt)
	if err != nil || len(result) == 0 {
		err := fmt.Errorf("no select %s.", itemName)
		return nil, err
	}

	chosen := make([]Choosable, 0, len(result))
	for _, r := range result {
		if c, ok := r.(Choosable); ok {
			chosen = append(chosen, c)
		}
	}

	return chosen, nil
}

func ChooseEC2(region, msg string) ([]string, error) {
	instances, err := GetInstances(region)
	if err != nil {
		return nil, err
	}

	choiceSlice := make([]Choosable, 0, len(instances))
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

		c := &Choice{
			C: choice,
			V: id,
		}

		choiceSlice = append(choiceSlice, c)
	}

	chosen, err := Choose("EC2 instance", "please select EC instance", choiceSlice)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(chosen))
	for _, c := range chosen {
		ids = append(ids, c.Value())
	}

	return ids, nil
}
