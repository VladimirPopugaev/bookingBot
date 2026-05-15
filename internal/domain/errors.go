package domain

import "errors"

var (
	ErrEmptyParameter   = errors.New("parameter is empty")
	ErrContextCancelled = errors.New("context cancelled")
	ErrInvalidParameter = errors.New("parameter is invalid")

	ErrURLParse      = errors.New("url parse failed")
	ErrCollectStruct = errors.New("collect site structure failed")
	ErrParseStruct = errors.New("parse site structure failed")
)
