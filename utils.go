package litejob

import "github.com/satori/go.uuid"

func GUID() string {

	uuid := uuid.NewV4()
	return uuid.String()
}
