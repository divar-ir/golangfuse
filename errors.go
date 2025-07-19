package golangfuse

import "errors"

var AlreadyStartedErr = errors.New("already started")
var AlreadyShutdownErr = errors.New("already shutdown")
