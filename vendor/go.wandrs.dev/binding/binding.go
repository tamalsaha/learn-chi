// Copyright 2014 Martini Authors
// Copyright 2014 The Macaron Authors
// Copyright 2020 The Gitea Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package binding is a middleware that provides request data binding and validation for Chi.
package binding

import (
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"go.wandrs.dev/inject"
)

var validate = validator.New()

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Bind wraps up the functionality of the Form and Json middleware
// according to the Content-Type and verb of the request.
// A Content-Type is required for POST and PUT requests.
// Bind invokes the ErrorHandler middleware to bail out if errors
// occurred. If you want to perform your own error handling, use
// Form or Json middleware directly. An interface pointer can
// be added as a second argument in order to map the struct to
// a specific interface.
func Bind(req *http.Request, obj interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := req.Header.Get("Content-Type")
			if req.Method == http.MethodPost || req.Method == http.MethodPut || len(contentType) > 0 {
				switch {
				case strings.Contains(contentType, "form-urlencoded"):
					Form(req, obj)
				case strings.Contains(contentType, "multipart/form-data"):
					MultipartForm(req, obj)
				case strings.Contains(contentType, "json"):
					JSON(req, obj)
				default:
					var errors Errors
					if contentType == "" {
						errors.Add([]string{}, ERR_CONTENT_TYPE, "Empty Content-Type")
					} else {
						errors.Add([]string{}, ERR_CONTENT_TYPE, "Unsupported Content-Type")
					}

					// handle error
					errorHandler(errors, w)
				}
			} else {
				Form(req, obj)
			}
		})
	}
}

const (
	_JSON_CONTENT_TYPE          = "application/json; charset=utf-8"
	STATUS_UNPROCESSABLE_ENTITY = 422
)

// errorHandler simply counts the number of errors in the
// context and, if more than 0, writes a response with an
// error code and a JSON payload describing the errors.
// The response will have a JSON content-type.
// Middleware remaining on the stack will not even see the request
// if, by this point, there are any errors.
// This is a "default" handler, of sorts, and you are
// welcome to use your own instead. The Bind middleware
// invokes this automatically for convenience.
func errorHandler(errs Errors, rw http.ResponseWriter) {
	if len(errs) > 0 {
		rw.Header().Set("Content-Type", _JSON_CONTENT_TYPE)
		if errs.Has(ERR_DESERIALIZATION) {
			rw.WriteHeader(http.StatusBadRequest)
		} else if errs.Has(ERR_CONTENT_TYPE) {
			rw.WriteHeader(http.StatusUnsupportedMediaType)
		} else {
			rw.WriteHeader(STATUS_UNPROCESSABLE_ENTITY)
		}
		errOutput, _ := json.Marshal(errs)
		rw.Write(errOutput)
		return
	}
}

// Form is middleware to deserialize form-urlencoded data from the request.
// It gets data from the form-urlencoded body, if present, or from the
// query string. It uses the http.Request.ParseForm() method
// to perform deserialization, then reflection is used to map each field
// into the struct with the proper type. Structs with primitive slice types
// (bool, float, int, string) can support deserialization of repeated form
// keys, for example: key=val1&key=val2&key=val3
// An interface pointer can be added as a second argument in order
// to map the struct to a specific interface.
func Form(formStruct interface{}, ifacePtr ...interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			var errors Errors

			ensureNotPointer(formStruct)
			formStruct := reflect.New(reflect.TypeOf(formStruct))
			parseErr := r.ParseForm()

			// Format validation of the request body or the URL would add considerable overhead,
			// and ParseForm does not complain when URL encoding is off.
			// Because an empty request body or url can also mean absence of all needed values,
			// it is not in all cases a bad request, so let's return 422.
			if parseErr != nil {
				errors.Add([]string{}, ERR_DESERIALIZATION, parseErr.Error())
			}

			// errors = mapForm(formStruct, r.Form, nil, errors)
			// validateAndMap(formStruct, ctx, errors, ifacePtr...)

			d := form.NewDecoder()
			if err := d.Decode(formStruct.Interface(), r.Form); err != nil {
				errors.Add([]string{}, "BAD_INPUT", err.Error())
			}

			if err := validate.Struct(formStruct.Interface()); err != nil {
				// errors = append(errors, err)
				// render error to the end user
			}

			//injector.Invoke(Validate(formStruct.Interface()))
			//errors = append(errors, getErrors(ctx)...)
			//injector.Map(errors)
			injector.Map(formStruct.Elem().Interface())
			if len(ifacePtr) > 0 {
				injector.MapTo(formStruct.Elem().Interface(), ifacePtr[0])
			}
		})
	}
}

// MaxMemory represents maximum amount of memory to use when parsing a multipart form.
// Set this to whatever value you prefer; default is 10 MB.
var MaxMemory = int64(1024 * 1024 * 10)

// MultipartForm works much like Form, except it can parse multipart forms
// and handle file uploads. Like the other deserialization middleware handlers,
// you can pass in an interface to make the interface available for injection
// into other handlers later.
func MultipartForm(formStruct interface{}, ifacePtr ...interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			var errors Errors
			ensureNotPointer(formStruct)
			formStruct := reflect.New(reflect.TypeOf(formStruct))
			// This if check is necessary due to https://github.com/martini-contrib/csrf/issues/6
			if r.MultipartForm == nil {
				// Workaround for multipart forms returning nil instead of an error
				// when content is not multipart; see https://code.google.com/p/go/issues/detail?id=6334
				if multipartReader, err := r.MultipartReader(); err != nil {
					errors.Add([]string{}, ERR_DESERIALIZATION, err.Error())
				} else {
					f, parseErr := multipartReader.ReadForm(MaxMemory)
					if parseErr != nil {
						errors.Add([]string{}, ERR_DESERIALIZATION, parseErr.Error())
					}

					if r.Form == nil {
						r.ParseForm()
					}
					for k, v := range f.Value {
						r.Form[k] = append(r.Form[k], v...)
					}

					r.MultipartForm = f
				}
			}
			// errors = mapForm(formStruct, r.MultipartForm.Value, r.MultipartForm.File, errors)
			// validateAndMap(formStruct, ctx, errors, ifacePtr...)

			d := form.NewDecoder()
			if err := d.Decode(formStruct.Interface(), r.Form); err != nil {
				errors.Add([]string{}, "BAD_INPUT", err.Error())
			}

			if err := validate.Struct(formStruct.Interface()); err != nil {
				// errors = append(errors, err)
				// render error to the end user
			}

			//injector.Invoke(Validate(formStruct.Interface()))
			//errors = append(errors, getErrors(injector)...)
			//injector.Map(errors)
			injector.Map(formStruct.Elem().Interface())
			if len(ifacePtr) > 0 {
				injector.MapTo(formStruct.Elem().Interface(), ifacePtr[0])
			}
		})
	}
}

// JSON is middleware to deserialize a JSON payload from the request
// into the struct that is passed in. The resulting struct is then
// validated, but no error handling is actually performed here.
// An interface pointer can be added as a second argument in order
// to map the struct to a specific interface.
//
// For all requests, Json parses the raw query from the URL using matching struct json tags.
//
// For POST, PUT, and PATCH requests, it also parses the request body.
// Request body parameters take precedence over URL query string values.
//
// Json follows the Request.ParseForm() method from Go's net/http library.
// ref: https://github.com/golang/go/blob/700e969d5b23732179ea86cfe67e8d1a0a1cc10a/src/net/http/request.go#L1176
func JSON(jsonStruct interface{}, ifacePtr ...interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			var errors Errors
			ensureNotPointer(jsonStruct)
			jsonStruct := reflect.New(reflect.TypeOf(jsonStruct))
			var err error
			if r.URL != nil {
				if params := r.URL.Query(); len(params) > 0 {
					d := form.NewDecoder()
					d.SetTagName("json")
					err = d.Decode(jsonStruct.Interface(), params)
				}
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				if r.Body != nil {
					v := jsonStruct.Interface()
					e := json.NewDecoder(r.Body).Decode(v)
					if err == nil {
						err = e
					}
				}
			}
			if err != nil && err != io.EOF {
				errors.Add([]string{}, ERR_DESERIALIZATION, err.Error())
			}

			if err = validate.Struct(jsonStruct.Interface()); err != nil {
				// errors = append(errors, err)
				// render error to the end user
			}

			//validateAndMap(jsonStruct, ctx, errors, ifacePtr...)
			//ctx.Invoke(Validate(jsonStruct.Interface()))
			//errors = append(errors, getErrors(ctx)...)
			//injector.Map(errors)
			injector.Map(jsonStruct.Elem().Interface())
			if len(ifacePtr) > 0 {
				injector.MapTo(jsonStruct.Elem().Interface(), ifacePtr[0])
			}
		})
	}
}

// Don't pass in pointers to bind to. Can lead to bugs.
func ensureNotPointer(obj interface{}) {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		panic("Pointers are not accepted as binding models")
	}
}

// Pointers must be bind to.
func ensurePointer(obj interface{}) {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		panic("Pointers are only accepted as binding models")
	}
}

type (
	// ErrorHandler is the interface that has custom error handling process.
	ErrorHandler interface {
		// Error handles validation errors with custom process.
		Error(*http.Request, Errors)
	}

	// Validator is the interface that handles some rudimentary
	// request validation logic so your application doesn't have to.
	Validator interface {
		// Validate validates that the request is OK. It is recommended
		// that validation be limited to checking values for syntax and
		// semantics, enough to know that you can make sense of the request
		// in your application. For example, you might verify that a credit
		// card number matches a valid pattern, but you probably wouldn't
		// perform an actual credit card authorization here.
		Validate(*http.Request, Errors) Errors
	}
)
