package domain

import "errors"

var (
	ErrEmptyParameter = errors.New("parameter is empty")
	ErrContextCancelled = errors.New("context cancelled")

	ErrURLParse = errors.New("url parse failed")
	ErrCollectStruct = errors.New("collect site structure failed")
)