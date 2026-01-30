package api

import "github.com/uug-ai/models/pkg/models"

// MarkerStatus represents specific status codes for marker operations
type MarkerStatus string

const (
	MarkerBindingFailed    MarkerStatus = "marker_binding_failed"
	MarkerDuplicateName    MarkerStatus = "marker_duplicate_name"
	MarkerMissingInfo      MarkerStatus = "marker_missing_info"
	MarkerFound            MarkerStatus = "marker_found"
	MarkerNotFound         MarkerStatus = "marker_not_found"
	MarkerAddSuccess       MarkerStatus = "marker_add_success"
	MarkerAddFailed        MarkerStatus = "marker_add_failed"
	MarkerUpdateSuccess    MarkerStatus = "marker_update_success"
	MarkerUpdateFailed     MarkerStatus = "marker_update_failed"
	MarkerDeleteSuccess    MarkerStatus = "marker_delete_success"
	MarkerDeleteFailed     MarkerStatus = "marker_delete_failed"
	MarkerRetrievalSuccess MarkerStatus = "marker_retrieval_success"
	MarkerRetrievalFailed  MarkerStatus = "marker_retrieval_failed"
	MarkerValidationFailed MarkerStatus = "marker_validation_failed"

	MarkerEventBindingFailed    MarkerStatus = "marker_event_binding_failed"
	MarkerEventRetrievalFailed  MarkerStatus = "marker_event_retrieval_failed"
	MarkerEventRetrievalSuccess MarkerStatus = "marker_event_retrieval_success"

	MarkerTagBindingFailed    MarkerStatus = "marker_tag_binding_failed"
	MarkerTagRetrievalFailed  MarkerStatus = "marker_tag_retrieval_failed"
	MarkerTagRetrievalSuccess MarkerStatus = "marker_tag_retrieval_success"

	MarkerCategoryBindingFailed    MarkerStatus = "marker_category_binding_failed"
	MarkerCategoryRetrievalFailed  MarkerStatus = "marker_category_retrieval_failed"
	MarkerCategoryRetrievalSuccess MarkerStatus = "marker_category_retrieval_success"
)

// String returns the string representation of the marker status
func (ms MarkerStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the marker status in the specified language
func (ms MarkerStatus) Translate(lang string) string {
	translations := map[string]map[MarkerStatus]string{
		"en": {
			MarkerBindingFailed:         "Marker binding failed",
			MarkerDuplicateName:         "Marker duplicate name",
			MarkerMissingInfo:           "Marker missing information",
			MarkerFound:                 "Marker found",
			MarkerNotFound:              "Marker not found",
			MarkerAddSuccess:            "Marker added successfully",
			MarkerAddFailed:             "Marker failed to add",
			MarkerUpdateSuccess:         "Marker updated successfully",
			MarkerUpdateFailed:          "Marker failed to update",
			MarkerDeleteSuccess:         "Marker deleted successfully",
			MarkerDeleteFailed:          "Marker failed to delete",
			MarkerRetrievalSuccess:      "Marker retrieved successfully",
			MarkerRetrievalFailed:       "Marker retrieval failed",
			MarkerValidationFailed:      "Marker validation failed",
			MarkerEventBindingFailed:    "Marker event binding failed",
			MarkerEventRetrievalFailed:  "Marker event retrieval failed",
			MarkerEventRetrievalSuccess: "Marker event retrieved successfully",
			MarkerTagBindingFailed:      "Marker tag binding failed",
			MarkerTagRetrievalFailed:    "Marker tag retrieval failed",
			MarkerTagRetrievalSuccess:   "Marker tag retrieved successfully",
		},
		"es": {
			MarkerBindingFailed:         "Error al vincular el marcador",
			MarkerDuplicateName:         "Nombre de marcador duplicado",
			MarkerMissingInfo:           "Información del marcador faltante",
			MarkerFound:                 "Marcador encontrado",
			MarkerNotFound:              "Marcador no encontrado",
			MarkerAddSuccess:            "Marcador agregado con éxito",
			MarkerAddFailed:             "Error al agregar el marcador",
			MarkerUpdateSuccess:         "Marcador actualizado con éxito",
			MarkerUpdateFailed:          "Error al actualizar el marcador",
			MarkerDeleteSuccess:         "Marcador eliminado con éxito",
			MarkerDeleteFailed:          "Error al eliminar el marcador",
			MarkerRetrievalSuccess:      "Marcador recuperado con éxito",
			MarkerRetrievalFailed:       "Error al recuperar el marcador",
			MarkerValidationFailed:      "Error de validación del marcador",
			MarkerEventBindingFailed:    "Error al vincular el evento del marcador",
			MarkerEventRetrievalFailed:  "Error al recuperar el evento del marcador",
			MarkerEventRetrievalSuccess: "Evento del marcador recuperado con éxito",
			MarkerTagBindingFailed:      "Error al vincular la etiqueta del marcador",
			MarkerTagRetrievalFailed:    "Error al recuperar la etiqueta del marcador",
			MarkerTagRetrievalSuccess:   "Etiqueta del marcador recuperada con éxito",
		},
		"fr": {
			MarkerBindingFailed:         "Échec de la liaison du marqueur",
			MarkerDuplicateName:         "Nom de marqueur dupliqué",
			MarkerMissingInfo:           "Informations manquantes sur le marqueur",
			MarkerFound:                 "Marqueur trouvé",
			MarkerNotFound:              "Marqueur non trouvé",
			MarkerAddSuccess:            "Marqueur ajouté avec succès",
			MarkerAddFailed:             "Échec de l'ajout du marqueur",
			MarkerUpdateSuccess:         "Marqueur mis à jour avec succès",
			MarkerUpdateFailed:          "Échec de la mise à jour du marqueur",
			MarkerDeleteSuccess:         "Marqueur supprimé avec succès",
			MarkerDeleteFailed:          "Échec de la suppression du marqueur",
			MarkerRetrievalSuccess:      "Marqueur récupéré avec succès",
			MarkerRetrievalFailed:       "Échec de la récupération du marqueur",
			MarkerValidationFailed:      "Échec de la validation du marqueur",
			MarkerEventBindingFailed:    "Échec de la liaison de l'événement du marqueur",
			MarkerEventRetrievalFailed:  "Échec de la récupération de l'événement du marqueur",
			MarkerEventRetrievalSuccess: "Événement du marqueur récupéré avec succès",
			MarkerTagBindingFailed:      "Échec de la liaison de la balise du marqueur",
			MarkerTagRetrievalFailed:    "Échec de la récupération de la balise du marqueur",
			MarkerTagRetrievalSuccess:   "Balise du marqueur récupérée avec succès",
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

type MarkerFilter struct {
	MarkerIds  []*string           `json:"markerIds,omitempty" bson:"markerIds,omitempty"`
	Names      []*string           `json:"names,omitempty" bson:"names,omitempty"`
	Name       *string             `json:"name,omitempty" bson:"name,omitempty"`
	Categories []*string           `json:"categories,omitempty" bson:"categories,omitempty"`
	DeviceKeys []*string           `json:"deviceKeys,omitempty" bson:"deviceKeys,omitempty"`
	TimeRanges []*models.TimeRange `json:"timeRanges,omitempty" bson:"timeRanges,omitempty"`
	Sort       *string             `json:"sort,omitempty" bson:"sort,omitempty"`
}

type MarkerEventFilter struct {
	MarkerEventIds []*string           `json:"markerEventIds,omitempty" bson:"markerEventIds,omitempty"`
	Names          []*string           `json:"names,omitempty" bson:"names,omitempty"`
	Name           *string             `json:"name,omitempty" bson:"name,omitempty"`
	DeviceKeys     []*string           `json:"deviceKeys,omitempty" bson:"deviceKeys,omitempty"`
	TimeRanges     []*models.TimeRange `json:"timeRanges,omitempty" bson:"timeRanges,omitempty"`
	Sort           *string             `json:"sort,omitempty" bson:"sort,omitempty"`
}

type MarkerTagFilter struct {
	Names      []*string           `json:"names,omitempty" bson:"names,omitempty"`
	Name       *string             `json:"name,omitempty" bson:"name,omitempty"`
	DeviceKeys []*string           `json:"deviceKeys,omitempty" bson:"deviceKeys,omitempty"`
	TimeRanges []*models.TimeRange `json:"timeRanges,omitempty" bson:"timeRanges,omitempty"`
	Sort       *string             `json:"sort,omitempty" bson:"sort,omitempty"`
}

type MarkerCategoryFilter struct {
	Names []*string `json:"names,omitempty" bson:"names,omitempty"`
	Name  *string   `json:"name,omitempty" bson:"name,omitempty"`
	Sort  *string   `json:"sort,omitempty" bson:"sort,omitempty"`
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

// GetMarkerOptions
// @Router /markers/options [post]
type GetMarkerOptionsRequest struct {
	Filter     *MarkerFilter     `json:"filter"`
	Pagination *CursorPagination `json:"pagination"`
}
type GetMarkerOptionsResponse struct {
	Markers []models.MarkerOption `json:"markers"`
}
type GetMarkerOptionsSuccessResponse struct {
	SuccessResponse
	Data GetMarkerOptionsResponse `json:"data"`
}
type GetMarkerOptionsErrorResponse struct {
	ErrorResponse
}

// GetMarkerEventOptions
// @Router /markers/events/options [post]
type GetMarkerEventOptionsRequest struct {
	Filter     *MarkerEventFilter `json:"filter"`
	Pagination *CursorPagination  `json:"pagination"`
}
type GetMarkerEventOptionsResponse struct {
	MarkerEvents []models.MarkerEventOption `json:"markerEvents"`
}
type GetMarkerEventOptionsSuccessResponse struct {
	SuccessResponse
	Data GetMarkerEventOptionsResponse `json:"data"`
}
type GetMarkerEventOptionsErrorResponse struct {
	ErrorResponse
}

// GetMarkerTagOptions
// @Router /markers/tags/options [post]
type GetMarkerTagOptionsRequest struct {
	Filter     *MarkerTagFilter  `json:"filter"`
	Pagination *CursorPagination `json:"pagination"`
}
type GetMarkerTagOptionsResponse struct {
	MarkerTags []models.MarkerTagOption `json:"markerTags"`
}
type GetMarkerTagOptionsSuccessResponse struct {
	SuccessResponse
	Data GetMarkerTagOptionsResponse `json:"data"`
}
type GetMarkerTagOptionsErrorResponse struct {
	ErrorResponse
}

// GetMarkerCategoryOptions
// @Router /markers/categories/options [post]
type GetMarkerCategoryOptionsRequest struct {
	Filter     *MarkerCategoryFilter `json:"filter"`
	Pagination *CursorPagination     `json:"pagination"`
}
type GetMarkerCategoryOptionsResponse struct {
	MarkerCategories []models.MarkerCategoryOption `json:"markerCategories"`
}
type GetMarkerCategoryOptionsSuccessResponse struct {
	SuccessResponse
	Data GetMarkerCategoryOptionsResponse `json:"data"`
}
type GetMarkerCategoryOptionsErrorResponse struct {
	ErrorResponse
}
