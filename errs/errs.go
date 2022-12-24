package errs

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"gopkg.in/go-playground/validator.v9"
)

func Error(err error) map[string]interface{} {
	result := make(map[string]interface{})
	if errCast, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errCast {
			result[strcase.ToLowerCamel(e.Field())] = validationErrorToText(e)
		}

		return result
	}

	result["error"] = err.Error()
	return result
}

func validationErrorToText(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", strcase.ToLowerCamel(e.Field()))
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", strcase.ToLowerCamel(e.Field()), e.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", strcase.ToLowerCamel(e.Field()), e.Param())
	case "email":
		return "invalid email format"
	case "len":
		return fmt.Sprintf("%s must be %s characters long", strcase.ToLowerCamel(e.Field()), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be %s", strcase.ToLowerCamel(e.Field()), e.Param())
	}

	return fmt.Sprintf("%s is not valid", strcase.ToLowerCamel(e.Field()))
}
