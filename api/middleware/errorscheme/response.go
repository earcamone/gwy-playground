package errorscheme

import (
	"encoding/json"
	"net/http"
	"os"
)

type errorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func errorResponseHandler(w http.ResponseWriter, err *AppError) {
	response := errorResponse{
		Status:  err.Code,
		Message: err.Message,
	}

	if err.Internal != nil && os.Getenv("ENV") == "develop" {
		response.Details = err.Internal.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Code)

	json.NewEncoder(w).Encode(response)
}
