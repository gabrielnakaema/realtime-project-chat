package validator

import "net/mail"

type Validator struct {
	Errors map[string][]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) Add(key, message string) {
	_, ok := v.Errors[key]
	if ok {
		v.Errors[key] = append(v.Errors[key], message)
		return
	}
	v.Errors[key] = []string{message}
}

func (v *Validator) Check(key string, message string, valid bool) {
	if valid {
		return
	}

	v.Add(key, message)
}

func NotBlank(value string) bool {
	return value != ""
}

func ValidEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}
	return true
}

func MinLength(value string, min int) bool {
	return len(value) >= min
}
