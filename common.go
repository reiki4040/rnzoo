package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
)

const (
	ENV_AWS_REGION = "AWS_REGION"
	ENV_HOME       = "HOME"

	RNZOO_DIR_NAME = ".rnzoo"

	OPT_SILENT  = "silent"
	OPT_VERBOSE = "verbose"
	OPT_REGION  = "region"

	OPT_INSTANCE_ID = "instance-id"
	OPT_EIP_ID      = "eip-id"

	OPT_FORCE           = "force"
	OPT_ALLOCATE        = "allocate"
	OPT_RELEASE         = "release"
	OPT_ASSOCIATE       = "associate"
	OPT_DISASSOCIATE    = "disassociate"
	OPT_WITHOUT_RELEASE = "without-release"
	OPT_REUSE           = "reuse"
	OPT_MOVE            = "move"
)

var silent bool
var verbose bool

func msg(v ...interface{}) {
	if !silent {
		log.Println(v...)
	}
}

func prepare(c *cli.Context) {
	silent = c.GlobalBool(OPT_SILENT)
	verbose = c.GlobalBool(OPT_VERBOSE)
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

func GetRnzooDir() string {
	rnzooDir := os.Getenv(ENV_HOME) + string(os.PathSeparator) + RNZOO_DIR_NAME
	return rnzooDir
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
