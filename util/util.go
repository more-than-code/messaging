package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
	"golang.org/x/crypto/bcrypt"
)

func IsEmail(phoneOrEmail string) bool {
	return strings.Contains(phoneOrEmail, "@")
}

func DomainFromAddress(address string) string {
	strGroups := strings.Split(address, "@")

	if len(strGroups) > 1 {
		return strGroups[1]
	}

	return ""
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
