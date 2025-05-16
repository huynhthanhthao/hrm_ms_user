package helper

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RespondWithError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

func ConvertIDsToUUIDStrings(data any) any {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)
			if field.Kind() == reflect.Slice && fieldType.Name == "Id" {
				bytesVal := field.Bytes()
				if u, err := uuid.FromBytes(bytesVal); err == nil {
					field.SetString(u.String())
				}
			} else if field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr {
				ConvertIDsToUUIDStrings(field.Interface())
			}
		}
	}
	return data
}
