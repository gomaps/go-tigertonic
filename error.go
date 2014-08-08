package tigertonic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	UnknownErrorType   = "unknown"
	UnknownErrorCode   = 0
	JSONErrorType      = "json"
	JSONErrorCode      = 9001
	MarshalerErrorType = "marshaler"
	MarshalerErrorCode = 9002
)

func NewMarshalerErrorEmptyInteface(method string) error {
	return &AppError{
		Type:           MarshalerErrorType,
		Code:           MarshalerErrorCode,
		Desc:           fmt.Sprintf("Empty interface is not suitable for %s request bodies", method),
		HttpStatusCode: http.StatusInternalServerError,
	}
}

func NewMarshalerErrorContentType(contentType string) error {
	return &AppError{
		Type:           MarshalerErrorType,
		Code:           MarshalerErrorCode,
		Desc:           fmt.Sprintf("Content-Type header is %s, not application/json", contentType),
		HttpStatusCode: http.StatusUnsupportedMediaType,
	}
}

func NewJSONError(desc string) error {
	return &AppError{
		Type:           JSONErrorType,
		Code:           JSONErrorCode,
		Desc:           desc,
		HttpStatusCode: http.StatusBadRequest,
	}
}

func NewMethodNotFoundError(desc string) error {
	return &AppError{
		Desc:           desc,
		HttpStatusCode: http.StatusNotFound,
	}
}

func NewMethodNotAllowed(desc string) error {
	return &AppError{
		Desc:           "Method not allowed, " + desc,
		HttpStatusCode: http.StatusMethodNotAllowed,
	}
}

func acceptJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if "" == accept {
		return true
	}
	return strings.Contains(accept, "*/*") || strings.Contains(accept, "application/json")
}

func errorName(err error, fallback string) string {
	if namedError, ok := err.(NamedError); ok {
		if name := namedError.Name(); "" != name {
			return name
		}
	}
	if httpEquivError, ok := err.(HTTPEquivError); ok && SnakeCaseHTTPEquivErrors {
		return strings.Replace(
			strings.ToLower(http.StatusText(httpEquivError.StatusCode())),
			" ",
			"_",
			-1,
		)
	}
	t := reflect.TypeOf(err)
	if reflect.Ptr == t.Kind() {
		t = t.Elem()
	}
	if r, _ := utf8.DecodeRuneInString(t.Name()); unicode.IsLower(r) {
		return fallback
	}
	return t.String()
}

func errorStatusCode(err error) int {
	// For pointers to AppError interface
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode()
	}

	// For direct interface to AppError
	if appErr, ok := err.(AppError); ok {
		return appErr.StatusCode()
	}

	// For direct interface to HTTPEquiv
	if httpEquivError, ok := err.(HTTPEquivError); ok {
		return httpEquivError.StatusCode()
	}

	return http.StatusInternalServerError
}

type AppErrorWrapper struct {
	Errors []error `json:"errors"`
}

type AppError struct {
	Type           string `json:"type,omitempty"`
	Code           int    `json:"code,omitempty"`
	Field          string `json:"field,omitempty"`
	Desc           string `json:"description,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (e AppError) Error() string {
	return e.Desc
}

func (e AppError) StatusCode() int {
	return e.HttpStatusCode
}

func (e AppError) ErrorType() string {
	return e.Type
}

func (e AppError) ErrorCode() int {
	return e.Code
}

func (e AppError) ErrorDesc() string {
	return e.Desc
}

func NewAppError(errCode int, errType string, errDesc string) *AppError {
	return &AppError{
		Code: errCode,
		Type: errType,
		Desc: errDesc,
	}
}

func WriteJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(errorStatusCode(err))

	//errs := []error{err}

	var errs []error

	if _, ok := err.(*AppError); ok {
		errs = []error{err}
	} else {
		errs = []error{AppError{
			Type: errorName(err, "error"),
			Desc: err.Error(),
		}}
	}

	jsonErrResponse := AppErrorWrapper{Errors: errs}

	if jsonErr := json.NewEncoder(w).Encode(jsonErrResponse); nil != jsonErr {
		log.Printf("Error marshalling error response into JSON output: %s", jsonErr)
	}
}

func WriteValidationErrors(w http.ResponseWriter, errs []error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	jsonErrResponse := AppErrorWrapper{Errors: errs}

	if jsonErr := json.NewEncoder(w).Encode(jsonErrResponse); nil != jsonErr {
		log.Printf("Error marshalling error response into JSON output: %s", jsonErr)
	}
}

func WritePlaintextError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorStatusCode(err))
	fmt.Fprintf(w, "%s: %s", errorName(err, "error"), err)
}
