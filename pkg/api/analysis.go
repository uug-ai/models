package api

import "github.com/uug-ai/models/pkg/models"

// AnalysisStatus represents specific status codes for analysis operations
type AnalysisStatus string

const (
	AnalysisFaceRedactionBindingFailed AnalysisStatus = "analysis_face_redaction_binding_failed"
	AnalysisSaveRedactionSuccess       AnalysisStatus = "analysis_save_redaction_success"
	AnalysisSaveRedactionFailed        AnalysisStatus = "analysis_save_redaction_failed"
	AnalysisSubmitRedactionSuccess     AnalysisStatus = "analysis_submit_redaction_success"
	AnalysisSubmitRedactionFailed      AnalysisStatus = "analysis_submit_redaction_failed"

	AnalysisFileNameMissing            AnalysisStatus = "analysis_file_name_missing"
	AnalysisSignedUrlMissing           AnalysisStatus = "analysis_signed_url_missing"
	AnalysisAllFrameCoordinatesMissing AnalysisStatus = "analysis_all_frame_coordinates_missing"

	AnalysisNotFound  AnalysisStatus = "analysis_not_found"
	AnalysisFound     AnalysisStatus = "analysis_found"
	AnalysisIdMissing AnalysisStatus = "analysisId_missing"
)

// String returns the string representation of the analysis status
func (as AnalysisStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the analysis status in the specified language
func (as AnalysisStatus) Translate(lang string) string {
	translations := map[string]map[AnalysisStatus]string{
		"en": {
			AnalysisFaceRedactionBindingFailed: "Face redaction binding failed",
			AnalysisSaveRedactionSuccess:       "Face redaction saved successfully",
			AnalysisSaveRedactionFailed:        "Failed to save face redaction",
			AnalysisSubmitRedactionSuccess:     "Face redaction submitted successfully",
			AnalysisSubmitRedactionFailed:      "Failed to submit face redaction",
			AnalysisSignedUrlMissing:           "Signed URL is missing",
			AnalysisNotFound:                   "Analysis not found",
			AnalysisFound:                      "Analysis found",
			AnalysisIdMissing:                  "Analysis ID is missing",
			AnalysisAllFrameCoordinatesMissing: "All frame coordinates map are missing",
			AnalysisFileNameMissing:            "File name is missing",
		},
		"es": {
			AnalysisFaceRedactionBindingFailed: "Error al vincular la redacción facial",
			AnalysisSaveRedactionSuccess:       "Redacción facial guardada con éxito",
			AnalysisSaveRedactionFailed:        "Error al guardar la redacción facial",
			AnalysisSubmitRedactionSuccess:     "Redacción facial enviada con éxito",
			AnalysisSubmitRedactionFailed:      "Error al enviar la redacción facial",
			AnalysisSignedUrlMissing:           "Falta la URL firmada",
			AnalysisNotFound:                   "Análisis no encontrado",
			AnalysisFound:                      "Análisis encontrado",
			AnalysisIdMissing:                  "Falta el ID del análisis",
			AnalysisFileNameMissing:            "Falta el nombre del archivo",
			AnalysisAllFrameCoordinatesMissing: "Faltan el mapa de coordenadas de todos los fotogramas",
		},
		"fr": {
			AnalysisFaceRedactionBindingFailed: "Échec de la liaison de la rédaction faciale",
			AnalysisSaveRedactionSuccess:       "Rédaction du visage enregistrée avec succès",
			AnalysisSaveRedactionFailed:        "Échec de l'enregistrement de la rédaction du visage",
			AnalysisSubmitRedactionSuccess:     "Rédaction du visage soumise avec succès",
			AnalysisSubmitRedactionFailed:      "Échec de la soumission de la rédaction du visage",
			AnalysisSignedUrlMissing:           "URL signée manquante",
			AnalysisNotFound:                   "Analyse non trouvée",
			AnalysisFound:                      "Analyse trouvée",
			AnalysisIdMissing:                  "ID d'analyse manquant",
			AnalysisFileNameMissing:            "Nom de fichier manquant",
			AnalysisAllFrameCoordinatesMissing: "Map de toutes les coordonnées des images sont manquantes",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[as]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[as]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return as.String()
}

// SaveFaceRedaction
// @Router /analysis/save-face-redaction [patch]
type SaveFaceRedactionRequest struct {
	AnalysisId    string               `json:"analysisId"`
	FaceRedaction models.FaceRedaction `json:"faceRedaction"`
}
type SaveFaceRedactionResponse struct {
	AnalysisId    string               `json:"analysisId"`
	FaceRedaction models.FaceRedaction `json:"faceRedaction"`
}
type SaveFaceRedactionSuccessResponse struct {
	SuccessResponse
	Data SaveFaceRedactionResponse `json:"data"`
}
type SaveFaceRedactionErrorResponse struct {
	ErrorResponse
}

// SubmitFaceRedaction
// @Router /analysis/submit-face-redaction [post]
type SubmitFaceRedactionRequest struct {
	AnalysisId          string                        `json:"analysisId"`
	SignedUrl           string                        `json:"signedUrl"`
	FileName            string                        `json:"fileName"`
	AllFrameCoordinates map[string][]*models.TrackBox `json:"allFrameCoordinates,omitempty"`
	FaceRedaction       *models.FaceRedaction         `json:"faceRedaction,omitempty"`
}
type SubmitFaceRedactionResponse struct {
	AnalysisId    string               `json:"analysisId"`
	FaceRedaction models.FaceRedaction `json:"faceRedaction"`
	Status        AnalysisStatus       `json:"status"`
}
type SubmitFaceRedactionSuccessResponse struct {
	SuccessResponse
	Data SubmitFaceRedactionResponse `json:"data"`
}
type SubmitFaceRedactionErrorResponse struct {
	ErrorResponse
}

type FaceRedactionMessage struct {
	Events []string               `json:"events,omitempty"`
	User   models.User            `json:"user,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// GetAnalysis
// @Router /analysis [get]
type GetAnalysisRequest struct {
	AnalysisId string `json:"analysisId"`
}
type GetAnalysisResponse struct {
	Analysis models.Analysis `json:"analysis"`
}
type GetAnalysisSuccessResponse struct {
	SuccessResponse
	Data GetAnalysisResponse `json:"data"`
}
type GetAnalysisErrorResponse struct {
	ErrorResponse
}
