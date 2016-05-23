package response

import (
	"github.com/emicklei/go-restful"
)

type Response struct {
	Success bool        `json:"success"`
	Error   *Error      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteSuccess(resp *restful.Response) {
	success := NewResponse(true)
	success.Write(resp)
}

func WriteResponse(data interface{}, resp *restful.Response) {
	WriteResponseStatus(200, data, resp)
}

func WriteResponseStatus(status int, data interface{}, resp *restful.Response) {
	success := NewResponse(true)
	success.Data = data
	success.WriteStatus(status, resp)
}

func NewResponse(success bool) *Response {
	return &Response{Success: success}
}

func NewErrorResponse(code string) *Response {
	res := &Response{Success: false, Error: &Error{}}
	res.SetError(code)
	return res
}

func (r *Response) SetError(code string) {
	if len(code) > 0 {
		if r.Error == nil {
			r.Error = &Error{}
		}
		//		r.Error.Name = err.Error()
		r.Error.Code = code
		r.Error.ErrorMsg = ErrorMsg(code)
	}
}

func (r *Response) WriteStatus(status int, resp *restful.Response) {
	if r.Error != nil && r.Error.Code == "" {
		r.Error.Code = string(status)
	}
	resp.WriteHeader(status)
	r.Write(resp)
}

func (r *Response) Write(resp *restful.Response) {
	resp.WriteEntity(r)
}
