package clickhouse

import "errors"

var (
	ErrClickAlreadyExists      = errors.New("click already exists for this ad and client")
	ErrImpressionAlreadyExists = errors.New("impression already exists for this ad and client")
	ErrClickAdNotShown         = errors.New("cannot record click: ad was not shown to this client")
)
