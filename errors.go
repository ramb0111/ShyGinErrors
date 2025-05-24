package ShyGinErrors

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type GinErrors interface {
	ListAllErrors(model interface{}, err error) map[string]string
}

type ginErrors struct {
	errorMaps map[string]string
}

type ErrorResult struct {
	Field   string
	JsonTag string
	Message string
}

func NewShyGinErrors(errors map[string]string) GinErrors {
	return &ginErrors{
		errorMaps: errors,
	}
}

func (ge ginErrors) ListAllErrors(model interface{}, err error) map[string]string {
	errors := map[string]string{}
	fields := map[string]ErrorResult{}

	if _, ok := err.(validator.ValidationErrors); ok {
		// resolve all json tags for the struct
		types := reflect.TypeOf(model)
		values := reflect.ValueOf(model)

		for i := 0; i < types.NumField(); i++ {
			field := types.Field(i)
			// Only exported fields have PkgPath == ""
			if field.PkgPath != "" {
				continue // skip unexported fields
			}
			value := values.Field(i).Interface()
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Name
			}
			messageTag := field.Tag.Get("msg")
			msg := ge.getErrorMessage(messageTag)

			if field.Type.Kind() == reflect.Slice {
				if value != nil {
					for i := 0; i < reflect.ValueOf(value).Len(); i++ {
						mps := ge.ListAllErrors(reflect.ValueOf(value).Index(i).Interface(), err)
						for k, v := range mps {
							errors[k] = v
						}
					}
				}
			}

			if field.Type.Kind() == reflect.Struct {
				if value != nil {
					mps := ge.ListAllErrors(value, err)
					for k, v := range mps {
						errors[k] = v
					}
				}
			}

			fmt.Printf("%s: %v = %v, tag= %v\n", field.Name, field.Type, value, jsonTag)
			fields[field.Name] = ErrorResult{
				Field:   field.Name,
				JsonTag: jsonTag,
				Message: msg,
			}
		}

		for _, e := range err.(validator.ValidationErrors) {
			if field, ok := fields[e.Field()]; ok {
				if field.Message != "" {
					errors[field.JsonTag] = field.Message
				} else {
					errors[field.JsonTag] = e.Error()
				}
			}
		}
	}else{
		errors["0"] = err.Error()
	}

	return errors
}

func (ge ginErrors) getErrorMessage(key string) string {
	if value, ok := ge.errorMaps[key]; ok {
		return value
	}
	return key
}
