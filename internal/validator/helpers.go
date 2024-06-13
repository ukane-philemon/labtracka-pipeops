package validator

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/constraints"
)

var (
	RgxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	PassWordErrorMsg = "password must have at least 8 characters, one lower and upper case letters, one digit, one special character"
)

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MinRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func MaxRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func Between[T constraints.Ordered](value, min, max T) bool {
	return value >= min && value <= max
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func In[T comparable](value T, safelist ...T) bool {
	for i := range safelist {
		if value == safelist[i] {
			return true
		}
	}
	return false
}

func AllIn[T comparable](values []T, safelist ...T) bool {
	for i := range values {
		if !In(values[i], safelist...) {
			return false
		}
	}
	return true
}

func NotIn[T comparable](value T, blocklist ...T) bool {
	for i := range blocklist {
		if value == blocklist[i] {
			return false
		}
	}
	return true
}

func NoDuplicates[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

func IsEmail(value string) bool {
	if len(value) > 254 {
		return false
	}

	return RgxEmail.MatchString(value)
}

func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}

// IsValidPhoneNumber returns true if the provided string is a valid phone
// number.
func IsValidPhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) < 9 || len(phoneNumber) > 16 {
		return false
	}

	regEx := regexp.MustCompile(`^[+]*[(]{0,1}[0-9]{1,4}[)]{0,1}[-\s\./0-9]*$`)
	if !regEx.MatchString(phoneNumber) {
		return false
	}

	_, parsable := ParsePhoneNumber(phoneNumber)
	return parsable

	// minLength, maxLength := 11, 11 // regular Nigeria phone number without +
	// if strings.HasPrefix(phoneNumber, "+") {
	// 	phoneNumber = phoneNumber[1:]
	// 	minLength, maxLength = 5, 25 // may be a foreign number, 5-25 digits excluding the +
	// }
	// return len(phoneNumber) >= minLength && len(phoneNumber) <= maxLength && IsDigitsOnly(phoneNumber)
}

// ParsePhoneNumber checks if the phone number is parsable and returns the
// parsed number.
func ParsePhoneNumber(phoneNumber string) (string, bool) {
	hasPlusPrefix := strings.HasPrefix(phoneNumber, "+") || strings.HasPrefix(phoneNumber, "(+") // may be there is a preceding bracket.
	switch {
	case hasPlusPrefix:
		return phoneNumber, true
	case len(phoneNumber) == 11 && strings.HasPrefix(phoneNumber, "0"):
		return "+234" + phoneNumber[1:], true // default to Nigeria.
	case len(phoneNumber) > 11:
		hasNGNPrefix := strings.HasPrefix(phoneNumber, "234") && len(phoneNumber) == 13 || strings.HasPrefix(phoneNumber, "2340") && len(phoneNumber) == 14
		if hasNGNPrefix {
			return "+" + phoneNumber, true
		}
		hasNGNPrefixWithBrackets := strings.HasPrefix(phoneNumber, "(234)") && len(phoneNumber) == 15 || strings.HasPrefix(phoneNumber, "(2340)") && len(phoneNumber) == 16
		if hasNGNPrefixWithBrackets {
			return "(+" + phoneNumber[1:], true
		}
	case len(phoneNumber) == 10 || !strings.HasPrefix(phoneNumber, "0"):
		return "+234" + phoneNumber, true // default to Nigeria.
	}

	return phoneNumber, false
}

// IsPasswordValid checks if the length of customers password meets the standard
// security length and characters.
func IsPasswordValid(password string) bool {
	if len(password) < 8 || len(password) > 70 {
		return false
	}

	checks := []string{"[a-z]", "[A-Z]", "[0-9]", "[^\\d\\w]"}
	for _, check := range checks {
		t, _ := regexp.MatchString(check, password)
		if !t {
			return false
		}
	}
	return true
}

func AnyValueEmpty(values ...string) bool {
	for _, v := range values {
		if v == "" {
			return true
		}
	}
	return false
}
