package validator

import (
	"log"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitV1() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("mcusername", mcusername); err != nil {
			log.Fatal(err)
		}
	}
}

func mcusername(fl validator.FieldLevel) bool {
	username, ok := fl.Field().Interface().(string)
	if ok {
		reg := regexp.MustCompile("^[a-zA-Z0-9_]{3,16}$")
		return reg.MatchString(username)
	}
	return false
}
