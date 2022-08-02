package terraform

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"io"
	"os"
	"path/filepath"
	"serverless-terraform-rds/code/static"

	"strings"
)

type OssConfig struct {
	OssBucket string `json:"oss_bucket"`
	OssPrefix string `json:"oss_prefix"`
	OssRegion string `json:"oss_region"`
}
type Client struct {
	logger *logrus.Entry
	// from payload
	Variables  json.RawMessage
	invokeType int

	stop chan int

	// from header
	AccessKey     string
	SecretKey     string
	SecurityToken string

	// from env variables
	OssPrefix   string
	OssBucket   string
	OssRegion   string
	OssEndpoint string
}

func NewTerraformClient(variables json.RawMessage, logger *logrus.Entry, invokeType int, stop chan int) *Client {
	return &Client{
		logger:     logger,
		Variables:  variables,
		invokeType: invokeType,
		stop:       stop,
	}
}

func (t *Client) GetOSSAndSecret(c *gin.Context, ossConfig *OssConfig) {
	t.AccessKey = c.GetHeader("x-fc-access-key-id")
	t.SecretKey = c.GetHeader("x-fc-access-key-secret")
	t.SecurityToken = c.GetHeader("x-fc-security-token")

	t.OssPrefix = ossConfig.OssPrefix

	t.OssBucket = ossConfig.OssBucket

	t.OssRegion = ossConfig.OssRegion

	// @todo change to internal network
	t.OssEndpoint = fmt.Sprintf("oss-%s-internal.aliyuncs.com", t.OssRegion)
}

func (t *Client) Validate() string {
	if t.AccessKey == "" {
		t.logger.Error(errors.New("Can't get AccessKey"))
		return "Can't get AccessKey"
	}
	if t.SecretKey == "" {
		t.logger.Error(errors.New("Can't get SecretKey"))
		return "Can't get SecretKey"
	}
	if t.SecurityToken == "" {
		t.logger.Error(errors.New("Can't get SecurityToken"))
		return "Can't get SecurityToken"
	}
	if t.OssRegion == "" {
		t.logger.Error(errors.New("Can't get oss region"))
		return "Can't get oss region"
	}
	if t.OssPrefix == "" {
		t.logger.Error(errors.New("Can't get oss prefix"))
		return "Can't get oss prefix"
	}
	if t.OssBucket == "" {
		t.logger.Error(errors.New("Can't get oss bucket"))
		return "Can't get oss bucket"
	}
	return ""
}

func (t *Client) apply() (error, string) {
	vars := []string{"apply", "-auto-approve", "-lock=false", "-json", "-var-file=vars.json"}

	secrets := make([]string, 0)
	secrets = append(secrets, fmt.Sprintf("ALICLOUD_ACCESS_KEY=%s", t.AccessKey), fmt.Sprintf("ALICLOUD_SECRET_KEY=%s", t.SecretKey), fmt.Sprintf("ALICLOUD_SECURITY_TOKEN=%s", t.SecurityToken))

	output, err := Execute("terraform", vars, t.logger, t.stop, &secrets)
	if err != nil {
		// new error found, write the error and raw outputs back.
		t.logger.Error(errors.New(fmt.Sprintf("Error: %s, raw_output_all: %s", err, output)))
		return errors.New(fmt.Sprintf("Error: %s, raw_output_all: %s", err, output)), ""
	}
	// get terraform output
	err, starts, outputs := t.getOutput(output)
	if err != nil {
		return errors.New(fmt.Sprintf("Error: %s, raw_output_all: %s", err, output)), ""
	}

	if len(starts) == 0 {
		return nil, outputs
	} else {
		t.logger.Error(errors.New(fmt.Sprintf("Didn't complete apply job, please check manully.\nstarts: \n%s\ncomplele: \n%s", starts, outputs)))
		return errors.New(fmt.Sprintf("Didn't complete apply job, please check manully.\nstarts: \n%s\ncomplele: \n%s", starts, outputs)), ""
	}

}

func (t *Client) destroy() (error, string) {
	vars := []string{"destroy", "-auto-approve", "-lock=false", "-json", "-var-file=vars.json"}

	secrets := make([]string, 0)
	secrets = append(secrets, fmt.Sprintf("ALICLOUD_ACCESS_KEY=%s", t.AccessKey), fmt.Sprintf("ALICLOUD_SECRET_KEY=%s", t.SecretKey), fmt.Sprintf("ALICLOUD_SECURITY_TOKEN=%s", t.SecurityToken))

	output, err := Execute("terraform", vars, t.logger, t.stop, &secrets)
	if err != nil {
		// error found, write the error and raw outputs back.
		t.logger.Error(errors.New(fmt.Sprintf("Error: %s, raw_output_all: %s", err, output)))
		return errors.New(fmt.Sprintf("Error: %s, raw_output_all: %s", err, output)), ""
	}
	err, starts, outputs := t.getOutput(output)
	if err != nil {
		return err, ""
	}

	if len(starts) == 0 {
		return nil, "Completely deleted resources"
	} else {
		t.logger.Error(errors.New(fmt.Sprintf("Didn't complete destroy job, please check manully.\nstarts: \n%s\ncomplele: \n%s", starts, outputs)))
		return errors.New(fmt.Sprintf("Didn't complete destroy job, please check manully.\nstarts: \n%s\ncomplele: \n%s", starts, outputs)), ""
	}
}

func (t *Client) init() error {
	vars := []string{"init", "-reconfigure"}

	secrets := make([]string, 0)
	secrets = append(secrets, fmt.Sprintf("ALICLOUD_ACCESS_KEY=%s", t.AccessKey), fmt.Sprintf("ALICLOUD_SECRET_KEY=%s", t.SecretKey), fmt.Sprintf("ALICLOUD_SECURITY_TOKEN=%s", t.SecurityToken))

	output, err := Execute("terraform", vars, t.logger, t.stop, &secrets)
	if err != nil {
		t.logger.Error(errors.New(fmt.Sprintf("Error: %s, output: %s", err, output)))
		return errors.New(fmt.Sprintf("Error: %s, output: %s", err, output))
	}
	err = t.checkLog(output)
	if err != nil {
		return err
	}
	return nil
}

// find error in log
func (t *Client) checkLog(logs string) error {

	lines := strings.Split(logs, "\n")

	for i, line := range lines {
		if strings.Contains(line, "31mError:") {
			errMsg := strings.Join(lines[i:], "\n")
			return errors.New(errMsg)
		}

	}
	return nil
}

func (t *Client) Do() (string, error) {

	// save tf data、variables into file.
	if err := t.save(); err != nil {
		return "", err
	}

	if t.invokeType == static.Apply {
		return Apply(t)
	}
	if t.invokeType == static.Delete {
		return Destroy(t)
	}

	return "", errors.New("Invalid operation, only allow apply or destroy")
}

// getOutput get outputs of terraform job.
func (t *Client) getOutput(output string) (error, string, string) {
	err, outputWrapperList := t.jsonListStrToOutputWrapperList(output)
	if err != nil {
		return err, "", ""
	}
	// apply_complete from terraform log
	completeList := make([]OutputWrapper, 0)
	// apply_start from terraform log
	startList := make([]OutputWrapper, 0)
	// outputs from terraform log
	outputList := make([]OutputWrapper, 0)

	for _, outputWrapper := range *outputWrapperList {
		if outputWrapper.Type == static.ApplyStart {
			startList = append(startList, outputWrapper)
		}

		if outputWrapper.Type == static.ApplyComplete {
			completeList = append(completeList, outputWrapper)
		}

		if outputWrapper.Type == static.Outputs {
			outputList = append(outputList, outputWrapper)
		}

	}
	// apply_start will end with apply_complete one by one.
	if len(completeList) == len(startList) && len(outputList) > 0 {
		// last item of outputList is the real output of terraform job.(from terraform cli output)
		outputs, err := json.Marshal(outputList[len(outputList)-1])
		if err != nil {
			return err, "", ""
		}
		return nil, "", string(outputs)
	}

	// when completeList != startList, it means the container is frozen and terraform job is stopped.
	starts, err := json.Marshal(startList)
	if err != nil {
		return err, "", ""
	}
	completes, err := json.Marshal(completeList)
	if err != nil {
		return err, "", ""
	}

	return nil, string(starts), string(completes)
}

func (t *Client) save() error {
	// find current path
	if err := os.MkdirAll("data", os.ModePerm); err != nil {
		t.logger.Error(errors.Wrap(err, "Create directory error"))
		return errors.Wrap(err, "Create directory error")
	}
	path, err := filepath.Abs(".")
	if err != nil {
		t.logger.Error(errors.Wrap(err, "Find current path error"))
		return errors.Wrap(err, "Find current path error")
	}

	// write variables to vars.json
	f, err := os.Create(path + "/data/vars.json")
	if err != nil {
		t.logger.Error(errors.Wrap(err, "Create variables file error"))
		return errors.Wrap(err, "Create variables file error")
	}

	if t.Variables == nil || len(t.Variables) == 0 {
		_, err2 := f.WriteString("{}")
		if err2 != nil {
			t.logger.Error(errors.Wrap(err2, "Write variables file error"))
			return errors.Wrap(err2, "Write variables file error")
		}
	} else {
		_, err2 := f.Write(t.Variables)
		if err2 != nil {
			t.logger.Error(errors.Wrap(err2, "Write variables file error"))
			return errors.Wrap(err2, "Write variables file error")
		}
	}
	if err := f.Close(); err != nil {

		t.logger.Error(errors.Wrap(err, "Close variables file error"))
		return errors.Wrap(err, "Close variables file error")
	}

	// write oss backend to oss.tf
	f, err = os.Create(path + "/data/backend.tf")
	if err != nil {
		t.logger.Error(errors.Wrap(err, "Create backend file error"))
		return errors.Wrap(err, "Create backend file error")
	}
	backendHcl := t.createBackend()
	_, err = f.Write(backendHcl)
	if err != nil {
		t.logger.Error(errors.Wrap(err, "Write hcl error"))
		return errors.Wrap(err, "Write hcl error")
	}

	if err = f.Close(); err != nil {
		t.logger.Error(errors.Wrap(err, "Close backend file error"))
		return errors.Wrap(err, "Close backend file error")
	}

	return nil
}

func (t Client) createBackend() []byte {
	newFile := hclwrite.NewEmptyFile()
	rootBody := newFile.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", nil)
	barBody := terraformBlock.Body()
	ossBlock := barBody.AppendNewBlock("backend", []string{"oss"})
	ossBody := ossBlock.Body()
	ossBody.SetAttributeValue("bucket", cty.StringVal(t.OssBucket))
	ossBody.SetAttributeValue("prefix", cty.StringVal(t.OssPrefix))
	ossBody.SetAttributeValue("key", cty.StringVal("mysql.tfstate"))
	ossBody.SetAttributeValue("region", cty.StringVal(t.OssRegion))
	ossBody.SetAttributeValue("endpoint", cty.StringVal(t.OssEndpoint))
	return newFile.Bytes()
}

// OutputWrapper Terraform outputs in console.
type OutputWrapper struct {
	Level      interface{}     `json:"@level"`
	Message    interface{}     `json:"@message"`
	Timestamp  interface{}     `json:"@timestamp"`
	Hook       json.RawMessage `json:"hook"`
	Type       interface{}     `json:"type"`
	Changes    json.RawMessage `json:"changes"`
	Outputs    json.RawMessage `json:"outputs"`
	Diagnostic json.RawMessage `json:"diagnostic"`
}

// When error, add apply_complete message and error message into output message.
// todo 之后用这个函数细化 error 的内容
func (t *Client) getOutputWhenError(output string) (error, string) {
	err, outputWrapperList := t.jsonListStrToOutputWrapperList(output)
	if err != nil {
		return err, ""
	}
	errMessages := make([]OutputWrapper, 0)
	for _, outputWrapper := range *outputWrapperList {
		if outputWrapper.Level == static.Error {
			errMessages = append(errMessages, outputWrapper)
		}
		if outputWrapper.Type == static.ApplyError {
			errMessages = append(errMessages, outputWrapper)
		}
		if outputWrapper.Type == static.ApplyComplete {
			errMessages = append(errMessages, outputWrapper)
		}
	}

	errorAndCompleted, err := json.Marshal(errMessages)
	if err != nil {
		return err, ""
	}
	return nil, string(errorAndCompleted)
}

func (t *Client) jsonListStrToOutputWrapperList(jsonListStr string) (error, *[]OutputWrapper) {
	dec := json.NewDecoder(strings.NewReader(jsonListStr))

	var outputList []OutputWrapper
	for {
		var outputWrapper OutputWrapper
		err := dec.Decode(&outputWrapper)
		if err == io.EOF {
			// all done
			break
		}
		if err != nil {
			t.logger.Error(err)
			return err, nil
		}
		outputList = append(outputList, outputWrapper)
	}

	return nil, &outputList
}
