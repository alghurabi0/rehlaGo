package validator

import (
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	Errors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}

	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// return true if n is bigger than m
func BiggerThanInt(n, m int) bool {
	return n > m
}

func FileTypeAllowed(buf []byte, allowed map[string]bool) bool {
	fileType := http.DetectContentType(buf)
	return allowed[fileType]
}

func FileSize(handler *multipart.FileHeader, maxSize int64) bool {
	return handler.Size <= maxSize
}

func ValidPhoneNumber(phone string) bool {
	match, _ := regexp.MatchString(`^07[0-9]{9}$`, phone)
	return match
}

func Password(pwd string) bool {
	return len(pwd) >= 8
}
