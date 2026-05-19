package app

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var phoneRegexp = regexp.MustCompile(`^\+?\d{10,15}$`)

type codedError struct {
	code    string
	message string
}

func (e codedError) Error() string {
	return e.message
}

func errCode(code, message string) error {
	return codedError{code: code, message: message}
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(phone)
}

func validateClientFields(name, phone, email string, requireEmail bool) error {
	if len([]rune(strings.TrimSpace(name))) < 2 {
		return errCode("INVALID_NAME", "Имя должно быть не короче 2 символов")
	}
	if !phoneRegexp.MatchString(normalizePhone(phone)) {
		return errCode("INVALID_PHONE", "Телефон должен содержать от 10 до 15 цифр")
	}
	if strings.TrimSpace(email) == "" {
		if requireEmail {
			return errCode("INVALID_EMAIL", "Email обязателен")
		}
		return nil
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
		return errCode("INVALID_EMAIL", "Некорректный email")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return errCode("WEAK_PASSWORD", "Пароль должен быть не короче 6 символов")
	}
	return nil
}

func validationCode(err error) (string, bool) {
	var coded codedError
	if errors.As(err, &coded) {
		return coded.code, true
	}
	return "", false
}
