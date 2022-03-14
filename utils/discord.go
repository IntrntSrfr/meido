package utils

import (
	"strconv"
	"strings"
	"time"
)

// IDToTimestamp converts a discord ID to a timestamp
func IDToTimestamp(idStr string) time.Time {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return time.Now()
	}

	id = ((id >> 22) + 1420070400000) / 1000
	return time.Unix(id, 0)
}

func TrimUserID(id string) string {
	id = strings.TrimPrefix(id, "<@")
	id = strings.TrimPrefix(id, "!")
	return strings.TrimSuffix(id, ">")
}
