package domain

import "errors"

var (
	ErrEmptyParameter = errors.New("parameter is empty")

	ErrUrlParse = errors.New("url parse failed")
	ErrCollectStruct = errors.New("collect site structure failed")
)