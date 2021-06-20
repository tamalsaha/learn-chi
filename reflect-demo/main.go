package main

import (
	"context"
	"fmt"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	httpw "go.wandrs.dev/http"
	"go.wandrs.dev/inject"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type Person struct {
	Name string `json:"name"`
}

type errorWrapper interface {
	error
}

func h_no_return(ctx context.Context) {

}

func h_returns_error(ctx context.Context) error {
	return nil
}

var _ error = &fs.PathError{}

func h_returns_custom_error(ctx context.Context) *fs.PathError {
	return nil
}

func h_returns_custom_error_ptr(ctx context.Context) fs.PathError {
	return fs.PathError{}
}

func h_returns_custom_error_interface(ctx context.Context) errorWrapper {
	return nil
}

func h_returns_string(ctx context.Context) string {
	return "handler"
}

func h_returns_int(ctx context.Context) int {
	return 69
}

func h_returns_bool(ctx context.Context) bool {
	return true
}

func h_returns_byte_array(ctx context.Context) []byte {
	return []byte("handler")
}

func h_returns_string_array(ctx context.Context) []string {
	return []string{"handler"}
}

func h_returns_int_array(ctx context.Context) []int {
	return []int{69}
}

func h_returns_bool_array(ctx context.Context) []bool {
	return []bool{true}
}

func h_returns_struct(ctx context.Context) Person {
	return Person{Name: "John"}
}

func h_returns_slice(ctx context.Context) []Person {
	return []Person{
		{Name: "John"},
		{Name: "Jane"},
	}
}

func h_returns_struct_err(ctx context.Context) (Person, error) {
	return Person{Name: "John"}, nil
}

func h_returns_too_many_returns(ctx context.Context) (int, Person, error) {
	return http.StatusOK, Person{Name: "John"}, nil
}

func main_create_new_object() {
	a := Person{
		Name: "a",
	}
	fmt.Printf("%+v\n", a)

	pt := reflect.TypeOf(a)

	// creates new object of Pointer type
	b_ptr := reflect.New(pt).Interface().(*Person)
	b_ptr.Name = "b"
	fmt.Printf("%+v\n", b_ptr)

	// creates new object of Pointer type. We use .Elem() to get the value type
	b := reflect.New(pt).Elem().Interface().(Person) // .Interface().(*Person2)
	b.Name = "b"
	fmt.Printf("%+v\n", b)
}

var (
	errorType          = reflect.TypeOf((*error)(nil)).Elem()
	responseWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
)

func main() {
	fn := h_returns_custom_error_interface

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	binding.Injector(render.New())(binding.HandlerFunc(fn)).ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}

func main__() {
	var w http.ResponseWriter = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	injector := inject.New()
	injector.Map(w)
	injector.Map(req)
	injector.Map(req.Context())
	injector.Map(httpw.NewResponseWriter(w, req, render.New()))

	typ := reflect.TypeOf(h_returns_custom_error_interface) //h_returns_custom_error_ptr)
	fmt.Println(typ.String())

	if typ.Kind() == reflect.Func {
		switch typ.NumOut() {
		case 0:
			return // nothing more to check
		case 1:
			// var err error
			// fmt.Println(reflect.TypeOf((error)(nil)))

			fmt.Println(errorType.String())
			rtyp := typ.Out(0)
			fmt.Println(rtyp.Implements(errorType))
			fmt.Println(reflect.New(rtyp).Type().String())
			fmt.Println(reflect.New(rtyp).Type().Implements(errorType))

		case 2:
		default:
			panic(fmt.Sprintf("found %d return values, allow at most 2", typ.NumOut()))
		}
	}
}

func main_() {
	fn := h_returns_custom_error

	var w http.ResponseWriter = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	injector := inject.New()
	injector.Map(w)
	injector.Map(req)
	injector.Map(req.Context())
	injector.Map(httpw.NewResponseWriter(w, req, render.New()))

	vals, err := injector.Invoke(fn)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(vals))

	ww := injector.GetVal(responseWriterType).Interface().(httpw.ResponseWriter)
	fmt.Println(ww)

	typ := reflect.TypeOf(fn)
	if typ.Kind() == reflect.Func {
		switch typ.NumOut() {
		case 0:
			return // nothing more to check
		case 1:
			// var err error
			// fmt.Println(reflect.TypeOf((error)(nil)))

			fmt.Println(errorType.String())
			rtyp := typ.Out(0)
			fmt.Println(rtyp.Implements(errorType))
		case 2:
			rtyp := typ.Out(1)
			if !rtyp.Implements(errorType) {
				panic("2nd return value must implement error")
			}
		default:
			panic(fmt.Sprintf("found %d return values, allow at most 2", typ.NumOut()))
		}
	}
}
