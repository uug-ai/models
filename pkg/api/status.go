package api

import (
	"net/http"
)

// https://pkg.go.dev/net/http#pkg-constants
const (
	StatusOK                  int = http.StatusOK
	StatusCreated             int = http.StatusCreated
	StatusUnauthorized        int = http.StatusUnauthorized
	StatusBadRequest          int = http.StatusBadRequest
	StatusInternalServerError int = http.StatusInternalServerError
	StatusNotFound            int = http.StatusNotFound
)

// Custom status codes for specific operations
const (
	StatusGetSuccess      int = 100
	StatusGetFailed       int = 101
	StatusGetSuccessEmpty int = 102
	StatusAddSuccess      int = 200
	StatusAddFailed       int = 201
)
