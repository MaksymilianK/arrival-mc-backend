package web

import "errors"

var ErrBadData = errors.New("data is invalid")
var ErrConflict = errors.New("a conflict has occurred")
var ErrNotFound = errors.New("the resource has not been found")
var ErrAuth = errors.New("the player is not authenticated")
var ErrPerm = errors.New("the player does not have a required permission")
