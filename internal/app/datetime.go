package app

import (
	"time"

	"github.com/pkg/errors"
)

func CurrentDatetime() (string, error) {
	tz, err := time.LoadLocation("Local")
	if err != nil {
		return "", errors.Wrap(err, "failed to load local timezone")
	}

	currentTime := time.Now().In(tz)
	return currentTime.Format(time.RFC1123), nil
}
