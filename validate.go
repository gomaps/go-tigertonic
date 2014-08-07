// © 2013 Steve McCoy under the MIT license.

/*
Package validate provides a type for automatically validating the fields of structs.

Any fields tagged with the key "validate" will be validated via a user-defined list of functions.
For example:

	type X struct {
		A string `validate:"long"`
		B string `validate:"short"`
		C string `validate:"long,proper"`
		D string
	}

Multiple validators can be named in the tag by separating their names with commas.
The validators are defined in a map like so:

	vd := make(validate.V)
	vd["long"] = func(i interface{}) error {
		…
	}
	vd["short"] = func(i interface{}) error {
		…
	}
	…

When present in a field's tag, the Validate method passes to these functions the value in the field
and should return an error when the value is deemed invalid.

There is a reserved tag, "struct", which can be used to automatically validate a
struct field, either named or embedded. This may be combined with user-defined validators.

Reflection is used to access the tags and fields, so the usual caveats and limitations apply.
*/
package tigertonic

import (
	"fmt"
	"reflect"
	"strings"
)

// V is a map of tag names to validators.
type V map[string]func(interface{}) error

// Package level validator
var Validator = make(V)

var ErrorCodeValidation int
var ErrorValidation string

func SetValidationError(code int, desc string) {
	ErrorCodeValidation = code
	ErrorValidation = desc
}

// BadField is an error type containing a field name and associated error.
// This is the type returned from Validate.
type BadField struct {
	ErrorString string `json:"error"`
	ErrorCode   int    `json:"errorCode"`
	Field       string `json:"field"`
	Desc        string `json:"description"`
}

func (b BadField) Error() string {
	return fmt.Sprintf("field %s is invalid: %v", b.Field, b.Desc)
}

// Validate accepts a struct (or a pointer) and returns a list of errors for all
// fields that are invalid. If all fields are valid, or s is not a struct type,
// Validate returns nil.
//
// Fields that are not tagged or cannot be interfaced via reflection
// are skipped.
func (v V) Validate(s interface{}) []error {
	var val reflect.Value

	// If the interface is a reflect.Value, do nothing
	if reflect.TypeOf(s).String() == "reflect.Value" {
		val = s.(reflect.Value)
	} else {
		val = reflect.ValueOf(s)
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	var errs []error

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := val.Field(i)
		if !fv.CanInterface() {
			continue
		}
		val := fv.Interface()
		tag := f.Tag.Get("validate")
		if tag == "" {
			continue
		}
		vts := strings.Split(tag, ",")

		for _, vt := range vts {
			if vt == "struct" {
				errs2 := v.Validate(val)
				if len(errs2) > 0 {
					errs = append(errs, errs2...)
				}
				continue
			}

			vf := v[vt]
			if vf == nil {
				errs = append(errs, BadField{
					ErrorString: ErrorValidation,
					ErrorCode:   ErrorCodeValidation,
					Field:       f.Name,
					Desc:        fmt.Sprintf("undefined validator: %q", vt),
				})
				continue
			}
			if err := vf(val); err != nil {
				p := fmt.Sprintf("%s", err)
				errs = append(errs, BadField{
					ErrorString: ErrorValidation,
					ErrorCode:   ErrorCodeValidation,
					Field:       fieldName(&f),
					Desc:        p,
				})
			}
		}
	}

	return errs
} //

func fieldName(f *reflect.StructField) string {
	jsonTag := f.Tag.Get("json")
	if jsonTag != "" {
		jsonTags := strings.Split(jsonTag, ",")
		if jsonTags[0] != "" {
			return jsonTags[0]
		}
	}

	return f.Name
}
