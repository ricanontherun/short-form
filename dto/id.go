package dto

import uuid "github.com/satori/go.uuid"

var (
	LongIDLength  = len(uuid.NamespaceDNS.String())
	ShortIDLength = 8
)

func IsValidId(id string) bool {
	var idLen = len(id)

	if idLen == LongIDLength {
		if _, err := uuid.FromString(id); err != nil {
			return false
		}
	} else if idLen != ShortIDLength {
		return false
	}

	return true
}

