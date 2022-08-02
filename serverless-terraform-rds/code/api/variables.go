package api

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type Database struct {
	Name         string `json:"name"`
	CharacterSet string `json:"character_set"`
	Description  string `json:"description"`
}

type Variables struct {
	Databases                *[]Database `json:"databases,omitempty"`
	UseVpc                   *bool       `json:"use_vpc,omitempty"`
	Region                   *string     `json:"region,omitempty"`
	InstanceName             *string     `json:"instance_name,omitempty"`
	AccountName              *string     `json:"account_name,omitempty"`
	Password                 *string     `json:"password,omitempty"`
	AllocatePublicConnection *bool       `json:"allocate_public_connection,omitempty"`
	SecurityIps              *[]string   `json:"security_ips,omitempty"`
	Privilege                *string     `json:"privilege,omitempty"`
	VswitchId                *string     `json:"vswitch_id,omitempty"`
}

const (
	Databases                = "DATABASES"
	UseVpc                   = "USE_VPC"
	Region                   = "REGION"
	InstanceName             = "INSTANCE_NAME"
	AccountName              = "ACCOUNT_NAME"
	Password                 = "PASSWORD"
	AllocatePublicConnection = "ALLOCATE_PUBLIC_CONNECTION"
	SecurityIps              = "SECURITY_IPS"
	Privilege                = "PRIVILEGE"
	VswitchId                = "VSWITCH_ID"
	InvokeType               = "INVOKE_TYPE"
)

func (v *Variables) getVarsFromEnv(logger *logrus.Entry) error {
	err := v.addDatabases()

	if err != nil {
		logger.Error(err)
		return err
	}

	v.UseVpc, err = addBoolVar(os.Getenv(UseVpc))
	if err != nil {
		logger.Error(errors.Wrap(err, "Get UseVpc from env error"))
		return errors.Wrap(err, "Get UseVpc from env error")
	}
	v.Region = addStringVar(os.Getenv(Region))

	v.InstanceName = addStringVar(os.Getenv(InstanceName))

	v.AccountName = addStringVar(os.Getenv(AccountName))
	v.Password = addStringVar(os.Getenv(Password))

	v.AllocatePublicConnection, err = addBoolVar(os.Getenv(AllocatePublicConnection))
	if err != nil {
		logger.Error(errors.Wrap(err, "Get AllocatePublicConnection from env error"))
		return errors.Wrap(err, "Get AllocatePublicConnection from env error")
	}
	err = v.addSecurityIps()
	if err != nil {
		logger.Error(err)
		return err
	}
	v.Privilege = addStringVar(os.Getenv(Privilege))
	v.VswitchId = addStringVar(os.Getenv(VswitchId))
	return nil
}

func (v *Variables) addSecurityIps() error {
	str := os.Getenv(SecurityIps)
	if str != "" {
		var d []string
		err := json.Unmarshal([]byte(str), &d)
		if err != nil {
			return errors.Wrap(err, "Unmarshal SecurityIps from env error")
		}
		v.SecurityIps = &d
		return nil
	} else {
		v.SecurityIps = nil
		return nil
	}

}

func (v *Variables) addDatabases() error {
	str := os.Getenv(Databases)
	if str != "" {
		d := make([]Database, 0)
		err := json.Unmarshal([]byte(str), &d)
		if err != nil {
			return errors.Wrap(err, "Unmarshal databases from env error")
		}
		v.Databases = &d
		return nil
	} else {
		return errors.New("Get databases from env error")
	}

}
func addStringVar(str string) *string {

	if str != "" {
		return addString(str)
	} else {
		return nil
	}

}
func addBoolVar(str string) (*bool, error) {
	if str != "" {
		res, err := addBool(str)
		return res, err
	} else {
		// default false
		res := false
		return &res, nil
	}

}

func addString(str string) *string {
	return &str
}

func addBool(b string) (*bool, error) {
	var t bool
	if b == "true" {
		t = true
	} else if b == "false" {
		t = false
	} else {
		return nil, errors.New("Bool Variables can only have two states: true or false")
	}
	return &t, nil
}

func getInvokeType(logger *logrus.Entry) (int, error) {
	str := os.Getenv(InvokeType)
	if str == "" {
		logger.Error(errors.New("Get invokeType from env error"))
		return -1, errors.New("Get invokeType from env error")
	}

	invokeType, err := strconv.Atoi(str)
	if err != nil {
		logger.Error(errors.New("Get invokeType from env error"))
		return -1, errors.New("Get invokeType from env error")
	}
	if invokeType != 0 && invokeType != 1 {
		logger.Error(errors.New("Get invokeType from env error"))
		return -1, errors.New("Get invokeType from env error")
	}
	return invokeType, nil
}
