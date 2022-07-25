package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
)

type ResponseCode int64

const (
	CodeSuccess ResponseCode = 600 + iota
	CodeNamespaceNotFound
	CodeNamespaceNotSet
	CodePodNotFound
	CodePodNotSet
	CodeContainerNotFound
	CodeContainerNotSet
	CodeInvalidParam
	CodeInternalError
)

var codeMsgMap = map[ResponseCode]string{
	CodeSuccess:           "success",
	CodeNamespaceNotFound: "namespace not found",
	CodeNamespaceNotSet:   "namespace not set",
	CodePodNotFound:       "pod not found",
	CodePodNotSet:         "pod not set",
	CodeContainerNotFound: "container not found",
	CodeContainerNotSet:   "container not set",
	CodeInvalidParam:      "invalid parameters",
	CodeInternalError:     "internal server error",
}

func (c ResponseCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeInternalError]
	}
	return msg
}

type ResponseData struct {
	Code ResponseCode `json:"code"`
	Msg  interface{}  `json:"msg"`
	Data interface{}  `json:"data"`
}

func ResponseError(c *gin.Context, code ResponseCode) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  code.Msg(),
		Data: nil,
	})
}

func ResponseErrorWithMsg(c *gin.Context, code ResponseCode, msg interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &ResponseData{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}

// IsForbiddenError returns true if give error is http.StatusForbidden, false otherwise.
func IsForbiddenError(err error) bool {
	status, ok := err.(*errors.StatusError)
	if !ok {
		return false
	}
	return status.ErrStatus.Code == http.StatusForbidden
}

func IsNotFoundError(err error) bool {
	status, ok := err.(*errors.StatusError)
	if !ok {
		return false
	}
	return status.ErrStatus.Code == http.StatusNotFound
}
