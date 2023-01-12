package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/saransh-khobragade/golang-postgres-kubernetes-gRPC/db/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false
}
