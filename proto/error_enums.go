package proto

import "errors"

var ErrorEnums = struct {
	ErrFileNameCanNotBeEmpty error
	ErrFileSizeCanNotBeZero  error
	ErrFileTypeCanNotBeEmpty error
	ErrNoPrepareFileInfo     error
}{
	ErrFileNameCanNotBeEmpty: errors.New("file name can not be empty"),
	ErrFileSizeCanNotBeZero:  errors.New("file size can not be zero"),
	ErrFileTypeCanNotBeEmpty: errors.New("file type can not be empty"),
	ErrNoPrepareFileInfo:     errors.New("no prepare file info"),
}
