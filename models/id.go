package models

import uuid "github.com/satori/go.uuid"

var (
	IdLength      = len(uuid.NamespaceDNS.String())
	ShortIdLength = 8
)

func IsValidId(id string) bool {
	var idLen = len(id)
	var validLength = idLen == IdLength || idLen == ShortIdLength
	var validForm = true

	if idLen == IdLength {
		if _, err := uuid.FromString(id); err != nil {
			validForm = false
		}
	}

	return validLength && validForm
}

func NewId() string {
	return uuid.NewV4().String()
}
