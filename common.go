package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	ENV_AWS_REGION = "AWS_REGION"
	ENV_HOME       = "HOME"
	ENV_RNZOO_DIR  = "RNZOO_DIR"

	RNZOO_DIR_NAME = ".rnzoo"

	OPT_SILENT  = "silent"
	OPT_VERBOSE = "verbose"
	OPT_REGION  = "region"
	OPT_TSV     = "tsv"

	OPT_INSTANCE_ID = "instance-id"
	OPT_EIP_ID      = "eip-id"
	OPT_I_TYPE      = "type"
	OPT_START       = "start"

	OPT_FORCE           = "force"
	OPT_ALLOCATE        = "allocate"
	OPT_RELEASE         = "release"
	OPT_ASSOCIATE       = "associate"
	OPT_DISASSOCIATE    = "disassociate"
	OPT_WITHOUT_RELEASE = "without-release"
	OPT_REUSE           = "reuse"
	OPT_MOVE            = "move"

	OPT_AMI_ID       = "ami-id"
	OPT_SYMBOL       = "symbol"
	OPT_SKELETON     = "skeleton"
	OPT_DRYRUN       = "dry-run"
	OPT_CONFIRM      = "confirm"
	OPT_SPECIFY_NAME = "specify-name"

	OPT_EC2_ANY_STATE   = "ec2-any-state"
	OPT_EXECUTE         = "execute"
	OPT_WITHOUT_CONFIRM = "without-confirm"

	OPT_TAG_PAIRS       = "pairs"
	OPT_TAG_DELETE_KEYS = "delete-keys"
)

var silent bool
var verbose bool

func msg(v ...interface{}) {
	if !silent {
		log.Println(v...)
	}
}

func prepare(c *cli.Context) {
	silent = c.Bool(OPT_SILENT)
	verbose = c.Bool(OPT_VERBOSE)
}

func debug(v ...interface{}) {
	if verbose {
		msg(v...)
	}
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func convertNilString(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}

func validateInstanceId(id string) error {
	if id == "" {
		return errors.New("instance id is empty.")
	}

	if len(id) != 10 {
		return errors.New("instance id is 8 char")
	}

	if !strings.HasPrefix(id, "i-") {
		return errors.New("instance id starts with i-")
	}

	// TODO hex check

	return nil
}

func GetRnzooDir() (string, error) {
	if envDir := os.Getenv(ENV_RNZOO_DIR); envDir != "" {
		// replace ~ -> home dir
		if i := strings.Index(envDir, "~"); i == 0 {
			user, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("can not resolved RNZOO_DIR ~ : %s", err.Error())
			}
			envDir = user.HomeDir + string(os.PathSeparator) + envDir[1:]
		}

		return envDir, nil
	}

	rnzooDir := os.Getenv(ENV_HOME) + string(os.PathSeparator) + RNZOO_DIR_NAME
	return rnzooDir, nil
}

func CreateRnzooDir() error {
	rnzooDir, err := GetRnzooDir()
	if err != nil {
		return err
	}

	if _, err = os.Stat(rnzooDir); os.IsNotExist(err) {
		err = os.Mkdir(rnzooDir, 0700)
		if err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}

	return nil
}
