package litejob

import (
	"os"
	"fmt"
)

var random *os.File = nil

func Guid() string {

	if random == nil {
		random,_ = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	}

	var b [16]byte = [16]byte{0}
	random.Read(b[:])
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

