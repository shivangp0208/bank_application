package api

import "github.com/go-playground/validator/v10"

// this is the main interface for all custom made validators to follow
type RequestValidator interface {
	Validator() map[string]func(fl validator.FieldLevel) bool
}

// every custom made validator will be of type requestValidator which will
// implement the RequestValidator interface
type requestValidator struct{}

// this will give a new instance of the requets validator containing all custom
// validator which we will require while registering this custom validator in gin
func NewRequestValidator() RequestValidator {
	return &requestValidator{}
}

func (v requestValidator) Validator() map[string]func(fl validator.FieldLevel) bool {
	return map[string]func(fl validator.FieldLevel) bool{
		"currency": validCurrency,
	}
}
