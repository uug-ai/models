package api

type CursorPagination struct {
	// Request fields (sent by client)
	Cursor string `json:"cursor,omitempty" bson:"cursor,omitempty"`
	Limit  int64  `json:"limit,omitempty" bson:"limit,omitempty"`

	// IncludeTotal asks the server to also count everything the filter matches,
	// not just the page being returned, and report it as Total.
	//
	// It is opt-in because the count is a second query over the whole matched
	// set while the page itself stops at Limit, so its cost scales with the
	// tenant rather than with the response. Most callers page for display and
	// never read Total; making them pay for it is how a cheap list turns into
	// an expensive one, and it is why this defaults to off.
	//
	// A caller that omits it gets Total unset rather than zero-as-a-count:
	// Total is omitempty, so "not asked for" and "asked for, and the answer is
	// nought" are the same wire value. Callers that need to tell the two apart
	// must rely on having asked.
	IncludeTotal bool `json:"includeTotal,omitempty" bson:"includeTotal,omitempty"`

	// Response fields (returned by server)
	NextCursor string `json:"nextCursor,omitempty" bson:"nextCursor,omitempty"`
	PrevCursor string `json:"prevCursor,omitempty" bson:"prevCursor,omitempty"`
	HasMore    bool   `json:"hasMore" bson:"hasMore"`

	// Optional numbered pagination support
	Page     int64 `json:"page,omitempty" bson:"page,omitempty"`
	PageSize int64 `json:"pageSize,omitempty" bson:"pageSize,omitempty"`
	// Total is the size of the whole matched set, populated only when the
	// request set IncludeTotal. It is not the length of the page.
	Total int64 `json:"total,omitempty" bson:"total,omitempty"`
}

// PaginationRequest/response ?
