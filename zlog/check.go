package zlog

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func CheckFatal(err error) {
	if err != nil {
		pc, errFile, lineNo, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		errFunc := ""
		if ok && details != nil {
			errFunc = details.Name()
		}
		SLogger.Fatalw("FATAL: It was fatal error", "error", err, "error_func", errFunc, "error_file", errFile, "error_line_no", lineNo)
	}
}

func CheckFatalm(err error, msg string) {
	if err != nil {
		pc, errFile, lineNo, ok := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		errFunc := ""
		if ok && details != nil {
			errFunc = details.Name()
		}
		SLogger.Fatalw("FATAL: "+msg, "error", err, "error_func", errFunc, "error_file", errFile, "error_line_no", lineNo)
	}
}

type ErrorAPIResponse struct {
	Error        error         `json:"_"`
	ErrorStr     string        `json:"error"`
	Msg          string        `json:"msg"`
	StatusCode   int           `json:"status_code"`
	DebugDetails *DebugDetails `json:"debug,omitempty"`
	TimeStamp    time.Time     `json:"time"`
	RecoveryLog  bool          `json:"_"`
	Request      *RequestData  `json:"request,omitempty"`
}

type DebugDetails struct {
	Func  string `json:"func,omitempty"`
	File  string `json:"file,omitempty"`
	Line  int    `json:"line,omitempty"`
	Stack string `json:"err_stack,omitempty"`
}

type RequestData struct {
	Id      string      `json:"id"`
	URI     string      `json:"uri"`
	Host    string      `json:"host"`
	Headers http.Header `json:"headers"`
	Body    interface{} `json:"body"`
	Method  string      `json:"method"`
}

func CheckAndAbortAPIError(err error, c *gin.Context, statusCode int) {
	if err != nil {
		errApiResp := handleError(err, "There was an error.", statusCode, 3, true, c)
		c.Error(err)
		panic(errApiResp)
	}
}

func CheckAndAbortAPIErrorMsg(err error, msg string, c *gin.Context, statusCode int) {
	if err != nil {
		errApiResp := handleError(err, msg, statusCode, 3, true, c)
		c.Error(err)
		panic(errApiResp)
	}
}

func CheckAndAbortErrorMsg(err error, msg string) {
	if err != nil {
		errApiResp := handleError(err, msg, http.StatusInternalServerError, 3, true, nil)
		panic(errApiResp)
	}
}

func CheckAndAbortError(err error) {
	if err != nil {
		errApiResp := handleError(err, "There was an error.", http.StatusInternalServerError, 3, true, nil)
		panic(errApiResp)
	}
}

func handleError(err error, msg string, statusCode int, debugDepth int, stackTrace bool, c *gin.Context) (errApiResp *ErrorAPIResponse) {
	if err != nil {
		debugData := debugDetails(debugDepth, stackTrace)
		requestDetails := requestDetails(c)
		errApiResp = &ErrorAPIResponse{
			Error:       err,
			ErrorStr:    err.Error(),
			Msg:         "ERROR: " + msg,
			StatusCode:  statusCode,
			TimeStamp:   time.Now().UTC(),
			RecoveryLog: false,
		}
		if isDebugRequested(c) && debugData != nil {
			errApiResp.DebugDetails = debugData
		}
		if isRequestNeeded(c) && requestDetails != nil {
			errApiResp.Request = requestDetails
		}
		// SLogger.Errorw(errApiResp.Msg, "error", err)
		return errApiResp
	}
	return nil
}

func isDebugRequested(c *gin.Context) bool {
	return c.Query("err_debug") == "true"
}

func isRequestNeeded(c *gin.Context) bool {
	return c.Query("err_req") == "true"
}

func debugDetails(debugDepth int, stackTrace bool) (debugDetails *DebugDetails) {
	if debugDepth > 0 {
		pc, errFile, lineNo, ok := runtime.Caller(debugDepth)
		details := runtime.FuncForPC(pc)
		errFunc := ""
		if ok && details != nil {
			errFunc = details.Name()
		}
		debugDetails = &DebugDetails{
			Func: errFunc,
			File: errFile,
			Line: lineNo,
		}
		if stackTrace {
			debugDetails.Stack = string(debug.Stack())
		}
	}
	return
}

func requestDetails(c *gin.Context) (request *RequestData) {
	if c != nil {
		request = &RequestData{
			Id:      c.Request.Header.Get("X-Request-Id"),
			URI:     c.Request.RequestURI,
			Host:    c.Request.Host,
			Headers: c.Request.Header,
			Method:  c.Request.Method,
		}

		body, er := ioutil.ReadAll(c.Request.Body)
		if er != nil {
			SLogger.Errorw("ERROR: Failed to Read Request.Body", "error", er)
		}
		// c.Request.Body = ioutil.NopCloser(bytes.NewReader(body)) // To reinsert the Body content if done Before Request Handler in Middleware
		if len(body) > 0 {
			if body[0] == '{' {
				var result map[string]interface{}
				json.Unmarshal(body, &result)
				request.Body = result
			} else if body[0] == '[' {
				var results []map[string]interface{}
				json.Unmarshal([]byte(body), &results)
				request.Body = results
			} else {
				bodyString := string(body)
				request.Body = bodyString
			}
		}
	}
	return
}
