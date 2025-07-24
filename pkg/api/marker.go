package api

import "github.com/uug-ai/models/pkg/models"

// MarkerStatus represents specific status codes for marker operations
type MarkerStatus string

const (
	MarkerBindingFailed MarkerStatus = "marker_binding_failed"
	MarkerDuplicateName MarkerStatus = "marker_duplicate_name"
	MarkerMissingInfo   MarkerStatus = "marker_missing_info"
	MarkerFound         MarkerStatus = "marker_found"
	MarkerNotFound      MarkerStatus = "marker_not_found"
	MarkerAddSuccess    MarkerStatus = "marker_add_success"
	MarkerAddFailed     MarkerStatus = "marker_add_failed"
	MarkerUpdateSuccess MarkerStatus = "marker_update_success"
	MarkerUpdateFailed  MarkerStatus = "marker_update_failed"
	MarkerDeleteSuccess MarkerStatus = "marker_delete_success"
	MarkerDeleteFailed  MarkerStatus = "marker_delete_failed"
)

// String returns the string representation of the marker status
func (ms MarkerStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the marker status in the specified language
func (ms MarkerStatus) Translate(lang string) string {
	translations := map[string]map[MarkerStatus]string{
		"en": {
			MarkerBindingFailed: "Marker binding failed",
			MarkerDuplicateName: "Marker duplicate name",
			MarkerMissingInfo:   "Marker missing information",
			MarkerFound:         "Marker found",
			MarkerNotFound:      "Marker not found",
			MarkerAddSuccess:    "Marker added successfully",
			MarkerAddFailed:     "Marker failed to add",
			MarkerUpdateSuccess: "Marker updated successfully",
			MarkerUpdateFailed:  "Marker failed to update",
			MarkerDeleteSuccess: "Marker deleted successfully",
			MarkerDeleteFailed:  "Marker failed to delete",
		},
		"es": {
			MarkerBindingFailed: "Error al vincular el marcador",
			MarkerDuplicateName: "Nombre de marcador duplicado",
			MarkerMissingInfo:   "Información del marcador faltante",
			MarkerFound:         "Marcador encontrado",
			MarkerNotFound:      "Marcador no encontrado",
			MarkerAddSuccess:    "Marcador agregado con éxito",
			MarkerAddFailed:     "Error al agregar el marcador",
			MarkerUpdateSuccess: "Marcador actualizado con éxito",
			MarkerUpdateFailed:  "Error al actualizar el marcador",
			MarkerDeleteSuccess: "Marcador eliminado con éxito",
			MarkerDeleteFailed:  "Error al eliminar el marcador",
		},
		"fr": {
			MarkerBindingFailed: "Échec de la liaison du marqueur",
			MarkerDuplicateName: "Nom de marqueur dupliqué",
			MarkerMissingInfo:   "Informations manquantes sur le marqueur",
			MarkerFound:         "Marqueur trouvé",
			MarkerNotFound:      "Marqueur non trouvé",
			MarkerAddSuccess:    "Marqueur ajouté avec succès",
			MarkerAddFailed:     "Échec de l'ajout du marqueur",
			MarkerUpdateSuccess: "Marqueur mis à jour avec succès",
			MarkerUpdateFailed:  "Échec de la mise à jour du marqueur",
			MarkerDeleteSuccess: "Marqueur supprimé avec succès",
			MarkerDeleteFailed:  "Échec de la suppression du marqueur",
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

// GetMarkers
// @Router /markers [get]
type GetMarkersRequest struct {
}
type GetMarkersResponse struct {
	Markers []models.Marker `json:"markers"`
}
type GetMarkersSuccessResponse struct {
	SuccessResponse
	Data GetMarkersResponse `json:"data"`
}
type GetMarkersErrorResponse struct {
	ErrorResponse
}

// AddMarker
// @Router /markers [post]
type AddMarkerRequest struct {
	Marker models.Marker `json:"marker"`
}
type AddMarkerResponse struct {
	Marker models.Marker `json:"marker"`
}
type AddMarkerSuccessResponse struct {
	SuccessResponse
	Data AddMarkerResponse `json:"data"`
}
type AddMarkerErrorResponse struct {
	ErrorResponse
}
