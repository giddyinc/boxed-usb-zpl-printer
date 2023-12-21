package usbzplprinter

import "errors"

var (
	ErrorDeviceNotFound        = errors.New("Can not detect any USB printer")
	ErrorEndpointNotAccessable = errors.New("Can not access endpoint")
)
