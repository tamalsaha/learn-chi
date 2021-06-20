package main

import (
	"errors"
	"fmt"
	"io/fs"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/go-playground/form/v4"
)

// <form method="POST">
//   <input type="text" name="Name" value="joeybloggs"/>
//   <input type="text" name="Age" value="3"/>
//   <input type="text" name="Gender" value="Male"/>
//   <input type="text" name="Address[0].Name" value="26 Here Blvd."/>
//   <input type="text" name="Address[0].Phone" value="9(999)999-9999"/>
//   <input type="text" name="Address[1].Name" value="26 There Blvd."/>
//   <input type="text" name="Address[1].Phone" value="1(111)111-1111"/>
//   <input type="text" name="active" value="true"/>
//   <input type="text" name="MapExample[key]" value="value"/>
//   <input type="text" name="NestedMap[key][key]" value="value"/>
//   <input type="text" name="NestedArray[0][0]" value="value"/>
//   <input type="submit"/>
// </form>

// Address contains address information
type Address struct {
	Name  string
	Phone string
}

// User contains user information
type User struct {
	Name        string
	Age         uint8
	Gender      string
	Address     []Address
	Active      bool `form:"active"`
	MapExample  map[string]string
	NestedMap   map[string]map[string]string
	NestedArray [][]string
}

// use a single instance of Decoder, it caches struct info
var decoder *form.Decoder

func main__() {
	sr := ToAPIError(nil, nil)
	fmt.Println(sr)

	decoder = form.NewDecoder()
	decoder.SetTagName("json")
	// decoder.SetMode(form.ModeExplicit)

	// this simulates the results of http.Request's ParseForm() function
	values := parseForm()

	var user User

	// must pass a pointer
	err := decoder.Decode(&user, values)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("%#v\n", user)
}

// this simulates the results of http.Request's ParseForm() function
func parseForm() url.Values {
	return url.Values{
		"Name":                []string{"joeybloggs"},
		"Age":                 []string{"3"},
		"Gender":              []string{"Male"},
		"Address[0].Name":     []string{"26 Here Blvd."},
		"Address[0].Phone":    []string{"9(999)999-9999"},
		"Address[1].Name":     []string{"26 There Blvd."},
		"Address[1].Phone":    []string{"1(111)111-1111"},
		"active":              []string{"true"},
		"MapExample[key]":     []string{"value"},
		"NestedMap[key][key]": []string{"value"},
		"NestedArray[0][0]":   []string{"value"},
	}
}

func ToAPIError(err error, obj interface{}) *apierrors.StatusError {
	if err == nil {
		return &apierrors.StatusError{metav1.Status{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Status",
				APIVersion: "v1",
			},
			Status: metav1.StatusSuccess,
			Code:   http.StatusOK,
		}}
	}

	switch t := err.(type) {
	case form.DecodeErrors:
		ot := reflect.TypeOf(obj)
		if ot.Kind() == reflect.Ptr {
			ot = ot.Elem()
		}
		causes := make([]metav1.StatusCause, 0, len(t))
		for field, err := range t {
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseTypeFieldValueInvalid,
				Message: err.Error(),
				Field:   field,
			})
		}
		return &apierrors.StatusError{metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusBadRequest,
			Reason: metav1.StatusReasonBadRequest,
			Details: &metav1.StatusDetails{
				//Group:  qualifiedKind.Group,
				//Kind:   qualifiedKind.Kind,
				//Name:   name,
				Causes: causes,
			},
			Message: fmt.Sprintf("failed to decode into %s.%s", ot.PkgPath(), ot.Name()),
		}}
	case *form.InvalidDecoderError:
		return apierrors.NewInternalError(err)
	default:
		return apierrors.NewInternalError(err)
	}
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflect.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func main_3() {
	if _, err := os.Open("non-existing"); err != nil {
		var pathError *fs.PathError
		if As(err, &pathError) {
			fmt.Println("Failed at path:", pathError.Path)
		} else {
			fmt.Println(err)
		}
	}

}

func main___() {
	if _, err := os.Open("non-existing"); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("file does not exist")
		} else {
			fmt.Println(err)
		}
	}

}

func main() {
	print2(&User{})
}

func print2(obj interface{}) {
	ot := reflect.TypeOf(obj)
	if ot.Kind() == reflect.Ptr {
		fmt.Println(ot.String())
		ot = ot.Elem()
		fmt.Println(ot.PkgPath(), ot.Name())
		return
	}
	fmt.Println(ot.PkgPath(), ot.Name())
}
