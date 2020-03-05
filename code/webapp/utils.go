package webapp

import (
	"net/url"
	"errors"
)

// GetQueryParam gets a single parameter from the URL query string.
// Returns error if parameter doesn't exist or multiple parameters exist.
func GetQueryParam(url *url.URL, param string) (string, error) {
	vals, ok := url.Query()[param]
	if !ok {
		return "", errors.New("Missing parameter: %s" + param)
	}
	if 1 < len(vals) {
		return "", errors.New("Too many parameters: %s" + param)
	}
	return vals[0], nil
}
