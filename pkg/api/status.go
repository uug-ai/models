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
	ApplicationStatusSuccess         int = 0 // General success status (similar to Linux exit code 0)
	ApplicationStatusError           int = 1 // General error status (similar to Linux exit code 1)
	ApplicationStatusGetSuccess      int = 100
	ApplicationStatusGetFailed       int = 101
	ApplicationStatusGetSuccessEmpty int = 102
	ApplicationStatusAddSuccess      int = 200
	ApplicationStatusAddFailed       int = 201
	ApplicationStatusAddDuplicate    int = 202
)
