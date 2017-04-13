package group

import (
	"fmt"
)

// FormatFakeUserID returns a user id
func FormatFakeUserID(uid int) string {
	return fmt.Sprintf("iamafakeuser%d", uid)
}
