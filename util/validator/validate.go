package validator

import (
	v "gopkg.in/go-playground/validator.v9"
	"regexp"
)

// use a single instance of Validate, it caches struct info
var validate *v.Validate

type ParamError struct {
	Field string `json:"field"`
}

// GetValidate 获得 validator 单例
func GetValidate() *v.Validate {
	if validate == nil {
		validate = v.New()
		if err := validate.RegisterValidation("phone_number", checkPhoneNumber, false); err != nil {
			panic(err)
		}
		if err := validate.RegisterValidation("password", checkPassword, false); err != nil {
			panic(err)
		}
	}
	return validate
}

// Validate 校验 struct
func Validate(obj interface{}) v.ValidationErrors {
	err := GetValidate().Struct(obj)
	if err != nil {
		return err.(v.ValidationErrors)
	}
	return nil
}

// checkPassword 检查密码复杂度，4 选 3: 数字，大写字母，特殊符号，小写字母。且长度为 8 ~ 20.
func checkPassword(fl v.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	if len(s) < 8 || len(s) > 20 {
		return false
	}
	matched, err := regexp.MatchString(`^[a-zA-Z0-9~!@#$%^&*()_\-=+'",.;?\\\[\]<>/]+$`, s)
	if err != nil || !matched {
		return false
	}

	cnt := 0
	matched, err = regexp.MatchString(`[a-z]`, s)
	if err != nil {
		return false
	} else if matched {
		cnt += 1
	}
	matched, err = regexp.MatchString(`[A-Z]`, s)
	if err != nil {
		return false
	} else if matched {
		cnt += 1
	}
	matched, err = regexp.MatchString(`[0-9]`, s)
	if err != nil {
		return false
	} else if matched {
		cnt += 1
	}
	matched, err = regexp.MatchString(`[~!@#$%^&*()_\-=+'",.;?\\\[\]<>/]`, s)
	if err != nil {
		return false
	} else if matched {
		cnt += 1
	}

	if cnt < 3 {
		return false
	}
	return true
}

func checkPhoneNumber(fl v.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	matched, err := regexp.MatchString(`^\+(\d{1,3})[0-9\-]+$`, s)
	if err != nil {
		return false
	}
	return matched
}
