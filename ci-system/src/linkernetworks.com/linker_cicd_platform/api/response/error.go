package response

import (
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
)

type Error struct {
	Code string    `json:"code"`
	ErrorMsg string `json:"errormsg"`
}

func WriteError(code string, resp *restful.Response) {
	log.Printf("[mora][error] %v", NewError(code))

	// Set response status code
	status := http.StatusInternalServerError

	// Write error response
	WriteStatusError(status, code, resp)
}

func WriteStatusError(status int, code string, resp *restful.Response) {
	errResp := NewResponse(false)
	errResp.SetError(code)
	errResp.WriteStatus(status, resp)
}
