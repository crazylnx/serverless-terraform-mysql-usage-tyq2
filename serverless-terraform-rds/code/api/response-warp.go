package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	StatusFail = "FAILED"
	StatusOk   = "SUCCESS"
)

func RespondError(c *gin.Context, errCode int, errMsg string) {
	c.JSON(errCode, gin.H{
		"status":  StatusFail,
		"message": errMsg,
	})
}

func respondInternalError(c *gin.Context, errMsg string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"status":  StatusFail,
		"message": errMsg,
	})
}

func respondOk(c *gin.Context, result interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": StatusOk,
		"result": result,
	})
}
