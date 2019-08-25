package inwxclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const (
	// TestAPI is the URL to the test/staging API..
	TestAPI = "https://api.ote.domrobot.com/jsonrpc/"
	// ProdAPI is the URL to the live API.
	ProdAPI = "https://api.domrobot.com/jsonrpc/"
)

type request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     string      `json:"clTRID"`
	Lang   string      `json:"lang"`
}

// Response from the API.
type Response struct {
	// Code of the operation.
	// Success: 1000 >= code <= 2000
	// Error: code >= 2000
	// See the following link for a list of result codes:
	// https://www.inwx.com/es/help/apidoc/f/ch04.html
	Code int `json:"code"`
	// Message contains a human readable explanation of the response code.
	Message string `json:"msg"`
	// Data contains the actual response data.
	Data interface{} `json:"resData"`
	// Reason is an additional error message.
	Reason string `json:"reason"`
	// ReasonCode is an additional short error message tag.
	ReasonCode string `json:"reasonCode"`
	// SVTRID stands for Server Transaction Identifier and may be helpful if you contact our support team.
	SVTRID string `json:"svTRID"`
}

type rpcError struct {
	code int
	message string
}

type RPCError interface {
	Code() int

	Error() string
}

func mapToError(r *Response) RPCError {
	if r.Code < 2000 {
		return nil
	}
	return &rpcError{
		code:    r.Code,
		message: r.Message,
	}
}

func (r *rpcError) Code() int {
	return r.code
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("(Code=%d) %q", e.code, e.message)
}

// DOMRobot is client for INWX's API.
type DOMRobot struct {
	apiURL *url.URL
	httpCl *http.Client
}

// NewDOMRobot creates a new DOMRobot.
func NewDOMRobot(apiURL string, defaultClient *http.Client) (*DOMRobot, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	cl, err := newClient(defaultClient)
	if err != nil {
		return nil, err
	}

	return &DOMRobot{
		apiURL: u,
		httpCl: cl,
	}, nil
}

// Do requests the RPC method with the given parameters.
func (jc *DOMRobot) Do(method string, params interface{}) (*Response, error) {
	msg, err := encode(method, params)
	if err != nil {
		return nil, err
	}
	resp, err := jc.httpCl.Post(jc.apiURL.String(), "application/json", msg)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rpcResp, err := decode(resp.Body)
	if err != nil {
		return nil, err
	}

	if method == "account.login" && rpcResp.Code == 1000 {
		jc.httpCl.Jar.SetCookies(jc.apiURL, resp.Cookies())
	}

	return rpcResp, mapToError(rpcResp)
}

func encode(method string, params interface{}) (io.Reader, error) {
	msg, err := json.Marshal(request{Method: method, ID: "github.com/klingtnet/inwxclient", Lang: "en", Params: params})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(msg), nil
}

func decode(body io.Reader) (*Response, error) {
	dec := json.NewDecoder(body)
	var resp Response
	err := dec.Decode(&resp)
	if err != nil {
		data, readErr := ioutil.ReadAll(body)
		if readErr != nil {
			return nil, errors.Wrap(err, readErr.Error())
		}
		return nil, errors.Wrapf(err, string(data))
	}
	return &resp, nil
}
