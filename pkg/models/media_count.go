package models

// MediaCount is exact below the requested limit and a lower bound when LimitReached is true.
type MediaCount struct {
	Count        int64 `json:"count"`
	LimitReached bool  `json:"limitReached"`
}
