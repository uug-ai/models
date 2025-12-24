package api

// MediaStatus represents specific status codes for media operations
type TaskStatus string

const (
	TaskBindingFailed TaskStatus = "Task_binding_failed"
	TaskDuplicateName TaskStatus = "Task_duplicate_name"
	TaskMissingInfo   TaskStatus = "Task_missing_info"
	TaskFound         TaskStatus = "Task_found"
	TaskNotFound      TaskStatus = "Task_not_found"
	TaskAddSuccess    TaskStatus = "Task_add_success"
	TaskAddFailed     TaskStatus = "Task_add_failed"
	TaskUpdateSuccess TaskStatus = "Task_update_success"
	TaskUpdateFailed  TaskStatus = "Task_update_failed"
	TaskDeleteSuccess TaskStatus = "Task_delete_success"
	TaskDeleteFailed  TaskStatus = "Task_delete_failed"
)

// String returns the string representation of the Task status
func (ms TaskStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the Task status in the specified language
func (ms TaskStatus) Translate(lang string) string {
	translations := map[string]map[TaskStatus]string{
		"en": {
			TaskBindingFailed: "Task binding failed",
			TaskDuplicateName: "Task duplicate name",
			TaskMissingInfo:   "Task missing information",
			TaskFound:         "Task found",
			TaskNotFound:      "Task not found",
			TaskAddSuccess:    "Task added successfully",
			TaskAddFailed:     "Task failed to add",
			TaskUpdateSuccess: "Task updated successfully",
			TaskUpdateFailed:  "Task failed to update",
			TaskDeleteSuccess: "Task deleted successfully",
			TaskDeleteFailed:  "Task failed to delete",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ms]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ms]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ms.String()
}
