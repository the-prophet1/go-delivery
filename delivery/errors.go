package delivery

/*
	Using Err... at the beginning can better let the IDE recognize the error code
*/

import "errors"

var (
	ErrReceiverBufferFulled = errors.New("channel is full,maybe not receiver or receiver too busy")

	ErrLocationDeleted = errors.New("location removed")
)
