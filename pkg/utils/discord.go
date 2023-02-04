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
		return time.Unix(0, 0)
	}

	id = ((id >> 22) + 1420070400000) / 1000
	return time.Unix(id, 0)
}

func TrimUserID(id string) string {
	id = strings.TrimPrefix(id, "<@")
	id = strings.TrimPrefix(id, "!")
	return strings.TrimSuffix(id, ">")
}

// IDToTimestamp2 takes a Discord snowflake and parses a timestamp from it
func IDToTimestamp2(idStr string) (time.Time, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(((id>>22)+1420070400000)/1000, 0), nil
}
