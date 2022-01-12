package v1

import (
	"Smoozzy/TicketSystemController/internal/cache"
	"Smoozzy/TicketSystemController/internal/model"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type CacheController struct {
	cache cache.Cache
	lg    *zap.Logger
}

func NewCacheController(cache cache.Cache, lg *zap.Logger) *CacheController {
	return &CacheController{cache: cache, lg: lg}
}

func (c *CacheController) CheckStatus(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		c.lg.Error("CheckStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	data.Status, err = c.cache.GetStatusFromCache(request.Context(), data.CustomerInternalID)
	err = json.NewEncoder(writer).Encode(&data)
	if err != nil {
		c.lg.Error("CheckStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}
