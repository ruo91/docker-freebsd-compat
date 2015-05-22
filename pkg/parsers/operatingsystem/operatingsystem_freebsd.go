// +build freebsd

package operatingsystem

import (
//	"bytes"
	"errors"
//	"io/ioutil"
)


func GetOperatingSystem() (string, error) {
	// TODO: Implement OS detection
	return "", errors.New("Unable to detect OS")
}

func IsContainerized() (bool, error) {
	// TODO: Implement jail detection
	return false, errors.New("Unable to check if we are in container")
}
