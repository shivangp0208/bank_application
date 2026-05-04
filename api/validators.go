package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shivangp0208/bank_application/util"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}

func checkForbiddenFields(c *gin.Context, forbiddenFields []string) error {

	byteData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("unable to read the req body for checking checking Forbidden Fields:%v", err)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(byteData))

	fieldMap := make(map[string]interface{})
	if err := json.Unmarshal(byteData, &fieldMap); err != nil {
		return nil
	}

	for _, field := range forbiddenFields {
		if _, exists := fieldMap[field]; exists {
			return fmt.Errorf("field '%s' is not allowed in this request", field)
		}
	}
	return nil
}
