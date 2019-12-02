package utils

import uuid "github.com/satori/go.uuid"

func MakeUUID() string {
	return uuid.Must(uuid.NewV4(), nil).String()
}
