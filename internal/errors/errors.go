package errors

import "errors"

var (
	ErrUserAlreadyExists                 = errors.New("user already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrInvalidOrderNumber                = errors.New("invalid order number")
	ErrInsufficientFunds                 = errors.New("insufficient funds")
	ErrOrderNotRegistered                = errors.New("order not registered")

	ErrInvalidTokenClaim    = errors.New("invalid token claim")
	ErrInvalidToken         = errors.New("invalid token")
	ErrUnexpectedSignMethod = errors.New("unexpected signing method")
)
