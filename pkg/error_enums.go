package pkg

import "errors"

var ErrorEnums = struct {
	ErrFileNameCanNotBeEmpty error
	ErrFileSizeCanNotBeZero  error
	ErrFileTypeCanNotBeEmpty error
	ErrNoPrepareFileInfo     error
	ErrFileNotExist          error
	ErrFileExist             error

	ErrBoxNotExist error

	ErrDepotNotExist error
}{
	ErrFileNameCanNotBeEmpty: errors.New("file name can not be empty"),
	ErrFileSizeCanNotBeZero:  errors.New("file size can not be zero"),
	ErrFileTypeCanNotBeEmpty: errors.New("file type can not be empty"),
	ErrNoPrepareFileInfo:     errors.New("no prepare file info"),
	ErrFileNotExist:          errors.New("file not exist"),
	ErrFileExist:             errors.New("file exist"),

	ErrBoxNotExist: errors.New("box not exist"),

	ErrDepotNotExist: errors.New("depot not exist"),
}
