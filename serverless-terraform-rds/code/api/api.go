package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"serverless-terraform-rds/code/internal/terraform"
	"time"

	"net/http"
)

type params struct {
	OssConfig  terraform.OssConfig `json:"oss_config"`
	Variables  json.RawMessage     `json:"variables"`
	InvokeType int                 `json:"invoke_type"`
}

func Invoke(c *gin.Context, logger *logrus.Entry, stop chan int) {

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	p := &params{}

	if err := json.Unmarshal(jsonData, p); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
	}

	client := terraform.NewTerraformClient(p.Variables, logger, p.InvokeType, stop)

	client.GetOSSAndSecret(c, &p.OssConfig)

	errMessage := client.Validate()
	if errMessage != "" {
		RespondError(c, http.StatusBadRequest, errMessage)
		return
	}
	outputs, err := client.Do()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOk(c, outputs)

}

func PreFreeze(logger *logrus.Entry, stop chan int, c *gin.Context) {
	if stop != nil {
		logger.Info("Begin to pre pause")
		// terraform needs two ctrl-c to stop.
		stop <- 0
		time.Sleep(5 * time.Second)
		stop <- 0
		return
	}
	logger.Info("No stop channel found, no need to stop")
	respondOk(c, "ok")
}
