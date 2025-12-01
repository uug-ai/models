package api

// StrategyStatus represents specific status codes for device operations
type StrategyStatus string

const (
	StrategyRetrievalSuccess StrategyStatus = "strategy_retrieval_success"
	StrategyBindingFailed    StrategyStatus = "strategy_binding_failed"
	StrategyDuplicateName    StrategyStatus = "strategy_duplicate_name"
	StrategyMissingInfo      StrategyStatus = "strategy_missing_info"
	StrategyRetrievalFailed  StrategyStatus = "strategy_retrieval_failed"
	StrategyFound            StrategyStatus = "strategy_found"
	StrategyNotFound         StrategyStatus = "strategy_not_found"
	StrategyAddSuccess       StrategyStatus = "strategy_add_success"
	StrategyAddFailed        StrategyStatus = "strategy_add_failed"
	StrategyUpdateSuccess    StrategyStatus = "strategy_update_success"
	StrategyUpdateFailed     StrategyStatus = "strategy_update_failed"
	StrategyDeleteSuccess    StrategyStatus = "strategy_delete_success"
	StrategyDeleteFailed     StrategyStatus = "strategy_delete_failed"
)

// String returns the string representation of the device status
func (ds StrategyStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the device status in the specified language
func (ds StrategyStatus) Translate(lang string) string {
	translations := map[string]map[StrategyStatus]string{
		"en": {
			StrategyBindingFailed:    "Strategy binding failed",
			StrategyDuplicateName:    "Strategy duplicate name",
			StrategyMissingInfo:      "Strategy missing information",
			StrategyRetrievalFailed:  "Strategy retrieval failed",
			StrategyFound:            "Strategy found",
			StrategyNotFound:         "Strategy not found",
			StrategyAddSuccess:       "Strategy added successfully",
			StrategyAddFailed:        "Strategy failed to add",
			StrategyUpdateSuccess:    "Strategy updated successfully",
			StrategyUpdateFailed:     "Strategy failed to update",
			StrategyDeleteSuccess:    "Strategy deleted successfully",
			StrategyDeleteFailed:     "Strategy failed to delete",
			StrategyRetrievalSuccess: "Strategy retrieved successfully",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ds]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ds]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ds.String()
}
