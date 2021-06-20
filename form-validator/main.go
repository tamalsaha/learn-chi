package main

import (
	"fmt"
	"github.com/go-playground/form/v4"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// User contains user information
type User struct {
	FirstName      string     `validate:"required"`
	LastName       string     `validate:"required"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `validate:"required,email"`
	FavouriteColor string     `validate:"iscolor"`                // alias for 'hexcolor|rgb|rgba|hsl|hsla'
	Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
}

// Address houses a users address information
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func main() {

	validate = validator.New()

	validateStruct()
	//validateVariable()
}

func validateStruct() {

	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		Phone:  "none",
	}

	user := &User{
		FirstName:      "Badger",
		LastName:       "Smith",
		Age:            135,
		Email:          "Badger.Smith@gmail.com",
		FavouriteColor: "#000-",
		Addresses:      []*Address{address},
	}

	// returns nil or ValidationErrors ( []FieldError )
	err := validate.Struct(user)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return
		}

		for _, err := range err.(validator.ValidationErrors) {

			//fmt.Println(err.Namespace())
			//fmt.Println(err.Field())
			//fmt.Println(err.StructNamespace())
			//fmt.Println(err.StructField())
			//fmt.Println(err.Tag())
			//fmt.Println(err.ActualTag())
			//fmt.Println(err.Kind())
			//fmt.Println(err.Type())
			//fmt.Println(err.Value())
			//fmt.Println(err.Param())
			fmt.Println(err.Error())
		}

		// from here you can create your own error messages in whatever language you wish
		return
	}

	// save user to database
}

func validateVariable() {

	myEmail := "joeybloggs.gmail.com"

	errs := validate.Var(myEmail, "required,email")

	if errs != nil {
		fmt.Println(errs) // output: Key: "" Error:Field validation for "" failed on the "email" tag
		return
	}

	// email ok, move on
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

	ot := reflect.TypeOf(obj)
	if ot.Kind() == reflect.Ptr {
		ot = ot.Elem()
	}

	switch t := err.(type) {
	case *validator.InvalidValidationError:
		return &apierrors.StatusError{metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusUnprocessableEntity,
			Reason: metav1.StatusReasonInvalid,
			//Details: &metav1.StatusDetails{
			//	Group:  qualifiedKind.Group,
			//	Kind:   qualifiedKind.Kind,
			//	Name:   name,
			//	Causes: causes,
			//},
			Message: err.Error(),
		}}
	case validator.ValidationErrors:
		causes := make([]metav1.StatusCause, 0, len(t))
		for i := range t {
			err := t[i]
			st := metav1.CauseTypeFieldValueInvalid
			if err.Tag() == "required" {
				st = metav1.CauseTypeFieldValueRequired
			}
			causes = append(causes, metav1.StatusCause{
				Type:    st,
				Message: err.Error(),
				Field:   err.Namespace(),
			})
		}
		return &apierrors.StatusError{metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusUnprocessableEntity,
			Reason: metav1.StatusReasonInvalid,
			Details: &metav1.StatusDetails{
				//Group:  qualifiedKind.Group,
				//Kind:   qualifiedKind.Kind,
				//Name:   name,
				Causes: causes,
			},
			// Message: fmt.Sprintf("%s %q is invalid: %v", qualifiedKind.String(), name, errs.ToAggregate()),
			Message: fmt.Sprintf("%s.%s is invalid", ot.PkgPath(), ot.Name()),
		}}
	case form.DecodeErrors:
		ot := reflect.TypeOf(obj)
		if ot.Kind() == reflect.Interface {
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
				// Group:  qualifiedKind.Group,
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
