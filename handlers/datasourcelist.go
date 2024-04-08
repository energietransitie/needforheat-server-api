package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/energietransitie/twomes-backoffice-api/internal/helpers"
	"github.com/energietransitie/twomes-backoffice-api/services"
	"github.com/energietransitie/twomes-backoffice-api/twomes/datasourcelist"
	"github.com/sirupsen/logrus"
)

type DataSourceListHandler struct {
	service *services.DataSourceListService
}

// Create a new DataSourceListHandler.
func NewDataSourceListHandler(service *services.DataSourceListService) *DataSourceListHandler {
	return &DataSourceListHandler{
		service: service,
	}
}

// Handle API endpoint for creating a new DataSourceList.
func (h *DataSourceListHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var request datasourcelist.DataSourceList
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return NewHandlerError(err, "bad request", http.StatusBadRequest).WithLevel(logrus.ErrorLevel)
	}

	DataSourceList, err := h.service.Create(request.Name, request.Items)

	if err != nil {
		if helpers.IsMySQLRecordNotFoundError(err) {
			return NewHandlerError(err, "not found", http.StatusNotFound)
		}

		if helpers.IsMySQLDuplicateError(err) {
			return NewHandlerError(err, "duplicate", http.StatusBadRequest)
		}

		return NewHandlerError(err, "internal server error", http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(DataSourceList)
	if err != nil {
		return NewHandlerError(err, "internal server error", http.StatusInternalServerError).WithLevel(logrus.ErrorLevel)
	}

	return nil
}
