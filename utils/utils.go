// Package utils provides public functions frequently used by other packages.
package utils

import (
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// RootDir returns the root of the application
func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

// Exists checks if a file exists or not
func Exists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func GetDomainFromURL(websiteURL string) (string, error) {
	if websiteURL == "" {
		return "", errors.New("websiteURL shouldn't be empty")
	}
	parsedUrl, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	hostname := strings.TrimPrefix(parsedUrl.Hostname(), "www.")
	if hostname == "" {
		return websiteURL, nil
	}
	return hostname, nil
}
