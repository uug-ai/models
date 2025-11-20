package api

type CursorPagination struct {
	// Request fields (sent by client)
	Cursor string `json:"cursor,omitempty" bson:"cursor,omitempty"`
	Limit  int64  `json:"limit,omitempty" bson:"limit,omitempty"`

	// Response fields (returned by server)
	NextCursor string `json:"nextCursor,omitempty" bson:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty" bson:"prevCursor,omitempty"`
	HasMore    bool   `json:"hasMore" bson:"hasMore"`

	// Optional numbered pagination support
	Page     int64 `json:"page,omitempty" bson:"page,omitempty"`
	PageSize int64 `json:"pageSize,omitempty" bson:"pageSize,omitempty"`
	Total    int64 `json:"total,omitempty" bson:"total,omitempty"`
}

// PaginationRequest/response ?
