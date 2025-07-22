package api

import (
	"net/http"
)

// https://pkg.go.dev/net/http#pkg-constants
const (
	HttpStatusOK                  int = http.StatusOK
	HttpStatusCreated             int = http.StatusCreated
	HttpStatusUnauthorized        int = http.StatusUnauthorized
	HttpStatusBadRequest          int = http.StatusBadRequest
	HttpStatusInternalServerError int = http.StatusInternalServerError
	HttpStatusNotFound            int = http.StatusNotFound
)

// Custom status codes for specific operations
const (
	ApplicationSuccess         int = 0 // General success status (similar to Linux exit code 0)
	ApplicationError           int = 1 // General error status (similar to Linux exit code 1)
	ApplicationGetSuccess      int = 100
	ApplicationGetFailed       int = 101
	ApplicationGetSuccessEmpty int = 102
	ApplicationAddSuccess      int = 200
	ApplicationAddFailed       int = 201
)
