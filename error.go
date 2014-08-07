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
	if httpEquivError, ok := err.(HTTPEquivError); ok {
		return httpEquivError.StatusCode()
	}
	return http.StatusInternalServerError
}

type JSONErrorResponse struct {
	Errors []error `json:"errors"`
}

type JSONError struct {
	ErrorStr  string `json:"error"`
	ErrorCode int    `json:"errorCode,omitempty"`
	Field     string `json:"field,omitempty"`
	Desc      string `json:"description,omitempty"`
}

func (e JSONError) Error() string {
	return fmt.Sprintf("%s (%d) - %s (%s)", e.ErrorStr, e.ErrorCode, e.Desc, e.Field)
}

func WriteJSONError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorStatusCode(err))

	var errs []error
	errs = append(errs, JSONError{
		ErrorStr: errorName(err, "error"),
		Desc:     err.Error(),
	})

	jsonErrResponse := JSONErrorResponse{Errors: errs}

	if jsonErr := json.NewEncoder(w).Encode(jsonErrResponse); nil != jsonErr {
		log.Printf("Error marshalling error response into JSON output: %s", jsonErr)
	}
}

func WriteValidationErrors(w http.ResponseWriter, errs []error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	jsonErrResponse := JSONErrorResponse{Errors: errs}

	if jsonErr := json.NewEncoder(w).Encode(jsonErrResponse); nil != jsonErr {
		log.Printf("Error marshalling error response into JSON output: %s", jsonErr)
	}
}

func WritePlaintextError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(errorStatusCode(err))
	fmt.Fprintf(w, "%s: %s", errorName(err, "error"), err)
}
