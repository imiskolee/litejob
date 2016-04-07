package litejob

import (
	"github.com/satori/go.uuid"
)


func Guid() string {

	uuid := uuid.NewV4()
	return uuid.String()
}
