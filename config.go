package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/reiki4040/cstore"
	"github.com/reiki4040/peco"
)

func doInit(c *cli.Context) error {
	m, err := NewCStoreManager()
	if err != nil {
		return cli.Exit(fmt.Sprintf("can not load EC2: %s", err.Error()), 1)
	}

	cs, err := m.New("config", cstore.TOML)
	if err != nil {
		return cli.Exit(fmt.Sprintf("error during init: %s", err.Error()), 1)
	}

	err = DoConfigWizard(cs)
	if err != nil {
		return cli.Exit(fmt.Sprintf("error during init: %s", err.Error()), 1)
	}

	return cli.Exit("saved rnzoo config.", 1)
}

func GetDefaultConfig() (*RnzooConfig, error) {
	// load config
	m, err := NewCStoreManager()
	if err != nil {
		return nil, err
	}

	cs, err := m.New("config", cstore.TOML)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = cs.Get(&config)
	if err != nil {
		return nil, err
	}

	return &config.Default, nil
}

type Config struct {
	Default RnzooConfig
}

func (c *Config) Validate() error {
	return c.Default.Validate()
}

type RnzooConfig struct {
	Name      string `toml:"profile_name,omitempty"`
	AWSRegion string `toml:"aws_region"`

	//AWSKey                     string `toml:"aws_access_key_id"`
	//AWSSecret                  string `toml:"aws_secret_access_key"`
}

func (c *RnzooConfig) Validate() error {
	// now no validation
	return nil
}

func DoConfigWizard(cs *cstore.CStore) error {
	chosenRegion, err := peco.Choose("AWS region", "Please select default AWS region", "", AWSRegionList)
	if err != nil {
		return fmt.Errorf("region choose error:%s", err.Error())
	}

	region := ""
	for _, c := range chosenRegion {
		region = c.Value()
		break
	}

	c := &Config{
		Default: RnzooConfig{
			AWSRegion: region,
		},
	}

	if err := cs.Save(c); err != nil {
		return err
	}

	return nil
}

var (
	AWSRegionList = []peco.Choosable{
		&peco.Choice{C: "ap-northeast-1 (Tokyo)", V: "ap-northeast-1"},
		&peco.Choice{C: "ap-southeast-1 (Singapore)", V: "ap-southeast-1"},
		&peco.Choice{C: "ap-southeast-2 (Sydney)", V: "ap-southeast-2"},
		&peco.Choice{C: "eu-central-1 (Frankfurt)", V: "eu-central-1"},
		&peco.Choice{C: "eu-west-1 (Ireland)", V: "eu-west-1"},
		&peco.Choice{C: "sa-east-1 (Sao Paulo)", V: "sa-east-1"},
		&peco.Choice{C: "us-east-1 (N. Virginia)", V: "us-east-1"},
		&peco.Choice{C: "us-west-1 (N. California)", V: "us-west-1"},
		&peco.Choice{C: "us-west-2 (Oregon)", V: "us-west-2"},
	}
)

func ask(msg, defaultValue string) (string, error) {
	fmt.Printf("%s[%s]:", msg, defaultValue)
	reader := bufio.NewReader(os.Stdin)

	ans, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("input err:%s", err.Error())
	}

	return ans, nil
}
