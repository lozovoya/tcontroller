package v1

import (
	"Smoozzy/TicketSystemController/internal/cache"
	"Smoozzy/TicketSystemController/internal/model"
	"Smoozzy/TicketSystemController/internal/ticketer"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var ErrIDChannelOperatorForBillingEmpty = errors.New("IDChannelOperatorForBilling is empty")
var ErrCustomerInternalIdEmpty = errors.New("CustomerInternalId is empty")
var ErrIDChannelOperatorEmpty = errors.New("IDChannelOperator is empty")
var ErrDescriptionEmpty = errors.New("Description is empty")
var ErrTTStartTimeEmpty = errors.New("TTStartTime is empty")
var ErrOperatorTTIdEmpty = errors.New("OperatorTTId is empty")

type Ticket struct {
	ticketer ticketer.Ticket
	cache    cache.Cache
	lg       *zap.Logger
}

func NewTicketer(ticketer ticketer.Ticket, cache cache.Cache, lg *zap.Logger) *Ticket {
	return &Ticket{ticketer: ticketer, cache: cache, lg: lg}
}

func (t *Ticket) CreateTicket(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("CreateTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.ticketer.IDChannelConverter(data.IDChannelOperator)
	if err != nil {
		t.lg.Error("CreateTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("CreateTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	cacheRecord := cache.CacheRecord{
		Source:                      data.Source,
		CustomerInternalID:          data.CustomerInternalID,
		IDChannelOperatorForBilling: billingID,
		IDChannelOperator:           data.IDChannelOperator,
		Description:                 data.Description,
		TTStartTimeTS:               data.StartTimeTS,
		TTStartTime:                 data.StartTime,
		TTClassification:            data.TTClassification,
		OperatorTTId:                data.OperatorTTId,
		FileName:                    data.FileName,
		File:                        data.File,
		Status:                      model.Creating,
		Created:                     time.Now().String(),
		Modified:                    time.Now().String(),
	}
	err = t.cache.WriteToCache(request.Context(), &cacheRecord)
	if err != nil {
		t.lg.Error("CreateTicket", zap.Error(err))
	}
	err = t.ticketer.CreateTicket(request.Context(), ticket)
	if err != nil {
		t.lg.Error("CreateTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}

func (t *Ticket) ReopenTicket(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("ReopenTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.cache.GetProcessingSystemFromCache(request.Context(), data.CustomerInternalID)
	if err != nil {
		t.lg.Error("ReopenTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("ReopenTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	cacheRecord := cache.CacheRecord{
		CustomerInternalID: data.CustomerInternalID,
		Status:             model.Working,
		Modified:           time.Now().String(),
	}
	err = t.cache.WriteToCache(request.Context(), &cacheRecord)
	if err != nil {
		t.lg.Error("ReopenTicket", zap.Error(err))
	}
	err = t.ticketer.ReopenTicket(request.Context(), ticket)
	if err != nil {
		t.lg.Error("ReopenTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}
func (t *Ticket) ChangeTicketStatus(writer http.ResponseWriter, request *http.Request) {
	//TODO метод требует доработки, нужно добавить какой новый статус делаем
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("ChangeTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.cache.GetProcessingSystemFromCache(request.Context(), data.CustomerInternalID)
	if err != nil {
		t.lg.Error("ChangeTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("ChangeTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = t.ticketer.ChangeTicketStatus(request.Context(), ticket)
	if err != nil {
		t.lg.Error("ChangeTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}
func (t *Ticket) CheckTicketStatus(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("CheckTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.cache.GetProcessingSystemFromCache(request.Context(), data.CustomerInternalID)
	if err != nil {
		t.lg.Error("CheckTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("CheckTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	err = t.ticketer.CheckTicketStatus(request.Context(), ticket)
	if err != nil {
		t.lg.Error("CheckTicketStatus", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}
func (t *Ticket) AddNoteToTicket(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("AddNoteToTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.cache.GetProcessingSystemFromCache(request.Context(), data.CustomerInternalID)
	if err != nil {
		t.lg.Error("AddNoteToTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("AddNoteToTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	cacheRecord := cache.CacheRecord{
		CustomerInternalID: data.CustomerInternalID,
		Modified:           time.Now().String(),
	}
	err = t.cache.WriteToCache(request.Context(), &cacheRecord)
	if err != nil {
		t.lg.Error("AddNoteToTicket", zap.Error(err))
	}
	err = t.ticketer.AddNoteToTicket(request.Context(), ticket)
	if err != nil {
		t.lg.Error("AddNoteToTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return
}
func (t *Ticket) CloseTicket(writer http.ResponseWriter, request *http.Request) {
	var data *model.TicketDTO
	err := json.NewDecoder(request.Body).Decode(&data)
	if err != nil {
		t.lg.Error("CloseTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	billingID, err := t.cache.GetProcessingSystemFromCache(request.Context(), data.CustomerInternalID)
	if err != nil {
		t.lg.Error("CloseTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ticket := t.makeTicket(data, data.MessageType, billingID)
	err = t.CheckInFields(data.MessageType, ticket)
	if err != nil {
		t.lg.Error("CloseTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	cacheRecord := cache.CacheRecord{
		CustomerInternalID: data.CustomerInternalID,
		Status:             model.Closed,
		Modified:           time.Now().String(),
	}
	err = t.cache.WriteToCache(request.Context(), &cacheRecord)
	if err != nil {
		t.lg.Error("CloseTicket", zap.Error(err))
	}
	err = t.ticketer.CloseTicket(request.Context(), ticket)
	if err != nil {
		t.lg.Error("CloseTicket", zap.Error(err))
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	return

}

func (t *Ticket) CheckInFields(method model.RequestType, data *model.Ticket) error {
	switch method {
	case model.Create:
		//if data.IDChannelOperatorForBilling == "" {
		//	return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorForBillingEmpty)
		//}
		if data.CustomerInternalId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrCustomerInternalIdEmpty)
		}
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.Description == "" {
			return fmt.Errorf("CheckInFields: %w", ErrDescriptionEmpty)
		}
		if data.TTStartTime == "" {
			return fmt.Errorf("CheckInFields: %w", ErrTTStartTimeEmpty)
		}
	case model.Status:
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.OperatorTTId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrOperatorTTIdEmpty)
		}
	case model.Close:
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.OperatorTTId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrOperatorTTIdEmpty)
		}
	case model.Reopen:
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.OperatorTTId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrOperatorTTIdEmpty)
		}
	case model.Wait:
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.OperatorTTId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrOperatorTTIdEmpty)
		}
	case model.Note:
		if data.IDChannelOperator == "" {
			return fmt.Errorf("CheckInFields: %w", ErrIDChannelOperatorEmpty)
		}
		if data.OperatorTTId == "" {
			return fmt.Errorf("CheckInFields: %w", ErrOperatorTTIdEmpty)
		}
	default:
		return fmt.Errorf("CheckInFields: no such method")

	}
	return nil
}

func (t *Ticket) makeTicket(data *model.TicketDTO, messageType model.RequestType, billingID string) *model.Ticket {
	var ticket = model.Ticket{
		MessageType:                 messageType,
		IDChannelOperatorForBilling: billingID,
		CustomerInternalId:          data.CustomerInternalID,
		IDChannelOperator:           data.IDChannelOperator,
		Description:                 data.Description,
		TTStartTimeTS:               data.StartTimeTS,
		TTStartTime:                 data.StartTime,
		TTClassification:            data.TTClassification,
		FileName:                    data.FileName,
		File:                        data.File,
		OperatorTTId:                data.OperatorTTId,
		EventTimestamp:              data.EventTimeTS,
		TimeStampString:             data.EventTime,
		TTStatus:                    data.Status,
		Comment:                     data.Comment,
		User:                        data.User,
	}

	return &ticket
}
