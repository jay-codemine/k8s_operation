package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
)

type Response struct {
	ctx *gin.Context
}

func NewResponse(c *gin.Context) *Response {
	return &Response{ctx: c}
}

type envelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
	// RequestID string `json:"request_id,omitempty"` // 可选：如果你有 request_id
}

func (r *Response) Success(data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	r.ctx.JSON(http.StatusOK, envelope{
		Code: 0,
		Msg:  "OK",
		Data: data,
	})
}

// 列表也统一塞进 data 里
func (r *Response) SuccessList(items interface{}, total interface{}) {
	r.ctx.JSON(http.StatusOK, envelope{
		Code: 0,
		Msg:  "OK",
		Data: gin.H{
			"list":  items,
			"total": total,
		},
	})
}

func (r *Response) ToErrorResponse(err *errorcode.Error) {
	c := r.ctx
	if err == nil {
		err = errorcode.InvalidParams
	}

	payload := gin.H{
		"code": err.Code(),
		"msg":  err.Msg(),
	}
	if d := err.Details(); len(d) > 0 {
		payload["details"] = d
	}

	c.AbortWithStatusJSON(err.StatusCode(), payload)
}
