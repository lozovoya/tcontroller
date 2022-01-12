package responseController

import (
	"TController/internal/cache"
	"TController/internal/model"
	"TController/internal/ticketer"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"go.uber.org/zap"
)

type receiver struct {
	out      chan *model.Ticket
	cache    cache.Cache
	ticketer ticketer.Ticket
	sources  map[string]string
	lg       *zap.Logger
}

func NewReceiver(out chan *model.Ticket,
	cache cache.Cache,
	ticketer ticketer.Ticket,
	lg *zap.Logger) Response {
	return &receiver{out: out, cache: cache, ticketer: ticketer, sources: make(map[string]string), lg: lg}
}

func (r *receiver) InitReceiversPull(n int) {
	for id := 1; id <= n; id++ {
		go r.ResponseReceiver(r.out, id)
		log.Printf("receiver %d is started", id)
	}
}

func (r *receiver) AddSource(name, uri string) {
	r.sources[name] = uri
	return
}

func (r *receiver) ResponseReceiver(out chan *model.Ticket, id int) {
	for message := range out {
		log.Printf("ResponseController.ResponseReceiver: got message by stream id %d: %v", id, message)
		ctx := context.Background()
		switch message.MessageType {
		case model.Create:
			r.CreateTicket(ctx, message)
		case model.Reopen:
			r.ReopenTicket(ctx, message)
		case model.Status:
			r.StatusTicket(ctx, message)
		case model.Note:
			r.NoteTicket(ctx, message)
		case model.Wait:
			r.WaitTicket(ctx, message)
		case model.Close:
			r.DoneTicket(ctx, message)
		default:
			r.lg.Error("ResponseController.ResponseReceiver: wrong message type")
		}
	}
}

func (r *receiver) CreateTicket(ctx context.Context, ticket *model.Ticket) {
	cacheRecord, err := r.cache.GetFromCacheByCustomerID(ctx, ticket.CustomerInternalId)
	if err != nil {
		r.lg.Error("ResponseController.CreateTicket", zap.Error(err))
		return
	}
	if cacheRecord.CustomerInternalID == "" {
		r.lg.Error("ResponseController.ResponseReceiver: no cache record")
		return
	}
	if ticket.TTStatus == "error" {
		if cacheRecord.Status == model.Error {
			if r.sources[cacheRecord.Source] != "" {
				r.lg.Info("Request was declined by all ticket systems") //todo добавить идентификатор запроса
				err = r.SendEvent(ctx, ticket, cacheRecord.Source)
				if err != nil {
					r.lg.Error("responseController.CreateTicket", zap.Error(err))
					return
				}
			}
			return
		}
		r.ReRouteTicket(ctx, cacheRecord)
		return
	}

	if cacheRecord.IDChannelOperatorForBilling != ticket.IDChannelOperatorForBilling {
		r.lg.Error("cacheRecord.IDChannelOperatorForBilling != ticket.IDChannelOperatorForBilling") //todo добавить идентификатор запроса
		return
	}
	cacheRecord.Status = model.Working
	//cacheRecord.IDChannelOperatorForBilling = ticket.IDChannelOperatorForBilling
	cacheRecord.OperatorTTId = ticket.OperatorTTId
	cacheRecord.Modified = time.Now().String()
	err = r.cache.WriteToCache(ctx, cacheRecord)
	if err != nil {
		r.lg.Error("ResponseController.CreateTicket", zap.Error(err))
		return
	}

	if r.sources[cacheRecord.Source] != "" {
		err = r.SendEvent(ctx, ticket, cacheRecord.Source)
		if err != nil {
			r.lg.Error("ResponseController.CreateTicket", zap.Error(err))
			return
		}
	}
	return
}

func (r *receiver) ReRouteTicket(ctx context.Context, cacheRecord *cache.CacheRecord) {
	cacheRecord.Status = model.Error
	cacheRecord.Modified = time.Now().String()
	cacheRecord.IDChannelOperatorForBilling = r.IDChannelConverter(cacheRecord.IDChannelOperator,
		cacheRecord.IDChannelOperatorForBilling)
	err := r.cache.WriteToCache(ctx, cacheRecord)
	if err != nil {
		r.lg.Error("ResponseController.ReRouteTicket", zap.Error(err))
		return
	}
	var ticket = model.Ticket{
		MessageType:                 model.Create,
		IDChannelOperatorForBilling: cacheRecord.IDChannelOperatorForBilling,
		CustomerInternalId:          cacheRecord.CustomerInternalID,
		IDChannelOperator:           cacheRecord.IDChannelOperator,
		Description:                 cacheRecord.Description,
		TTStartTime:                 cacheRecord.TTStartTime,
		TTClassification:            cacheRecord.TTClassification,
		FileName:                    cacheRecord.FileName,
		File:                        cacheRecord.File,
	}
	err = r.ticketer.CreateTicket(ctx, &ticket)
	if err != nil {
		r.lg.Error("ResponseController.ReRouteTicket", zap.Error(err))
		return
	}
	return
}

func (r *receiver) ReopenTicket(ctx context.Context, ticket *model.Ticket) {
	//todo явно нужно изменить статус тикета. на пути туда или будет ответ?
}

func (r *receiver) StatusTicket(ctx context.Context, ticket *model.Ticket) {
	cacheRecord, err := r.cache.GetFromCacheByCustomerID(ctx, ticket.CustomerInternalId)
	if err != nil {
		r.lg.Error("responseController.StatusTicket", zap.Error(err))
		return
	}
	if cacheRecord.CustomerInternalID == "" {
		r.lg.Error("responseController.StatusTicket: no cache record")
		return
	}
	if r.sources[cacheRecord.Source] != "" {
		err = r.SendEvent(ctx, ticket, cacheRecord.Source)
		if err != nil {
			r.lg.Error("responseController.StatusTicket", zap.Error(err))
			return
		}
	}
	return
}

func (r *receiver) NoteTicket(ctx context.Context, ticket *model.Ticket) {
	cacheRecord, err := r.cache.GetFromCacheByCustomerID(ctx, ticket.CustomerInternalId)
	if err != nil {
		r.lg.Error("responseController.NoteTicket", zap.Error(err))
		return
	}
	if cacheRecord.CustomerInternalID == "" {
		r.lg.Error("responseController.NoteTicket: no cache record")
		return
	}
	if r.sources[cacheRecord.Source] != "" {
		err = r.SendEvent(ctx, ticket, cacheRecord.Source)
		if err != nil {
			r.lg.Error("responseController.NoteTicket", zap.Error(err))
			return
		}
	}
	return
}

func (r *receiver) WaitTicket(ctx context.Context, ticket *model.Ticket) {
	cacheRecord, err := r.cache.GetFromCacheByCustomerID(ctx, ticket.CustomerInternalId)
	if err != nil {
		r.lg.Error("responseController.WaitTicket", zap.Error(err))
		return
	}
	if cacheRecord.CustomerInternalID == "" {
		r.lg.Error("responseController.WaitTicket: no cache record")
		return
	}
	if r.sources[cacheRecord.Source] != "" {
		err = r.SendEvent(ctx, ticket, cacheRecord.Source)
		if err != nil {
			r.lg.Error("responseController.WaitTicket", zap.Error(err))
			return
		}
	}
	return
}

func (r *receiver) DoneTicket(ctx context.Context, ticket *model.Ticket) {
	cacheRecord, err := r.cache.GetFromCacheByCustomerID(ctx, ticket.CustomerInternalId)
	if err != nil {
		r.lg.Error("responseController.DoneTicket", zap.Error(err))
		return
	}
	if cacheRecord.CustomerInternalID == "" {
		r.lg.Error("responseController.DoneTicket: no cache record")
		return
	}
	cacheRecord.Status = model.Closed
	cacheRecord.Modified = time.Now().String()
	err = r.cache.WriteToCache(ctx, cacheRecord)

	if r.sources[cacheRecord.Source] != "" {
		err = r.SendEvent(ctx, ticket, cacheRecord.Source)
		if err != nil {
			r.lg.Error("responseController.DoneTicket", zap.Error(err))
			return
		}
	}
	return
}

func (r *receiver) SendEvent(ctx context.Context, ticket *model.Ticket, source string) error {
	var data = model.TicketDTO{
		Source:                      source,
		MessageType:                 ticket.MessageType,
		CustomerInternalID:          ticket.CustomerInternalId,
		IDChannelOperatorForBilling: ticket.IDChannelOperatorForBilling,
		IDChannelOperator:           ticket.IDChannelOperator,
		Description:                 ticket.Description,
		StartTimeTS:                 ticket.TTStartTimeTS,
		EventTimeTS:                 ticket.EventTimestamp,
		TTClassification:            ticket.TTClassification,
		FileName:                    ticket.FileName,
		//File:                        base64.Encoding{},
		OperatorTTId: ticket.OperatorTTId,
		Status:       ticket.TTStatus, //todo матрица транслируемых в заказчика статусов? так и не согласовали
		Comment:      ticket.Comment,
		User:         ticket.User,
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("responseController.SendEvent: %w", err)
	}
	response, err := http.Post(r.sources[source], "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("responseController.SendEvent: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("responseController.SendEvent: %s", response.Status)
	}
	return nil
}

func (r *receiver) IDChannelConverter(idChannelOperator, idChannelOperatorForBilling string) (altIDChannelOperatorForBilling string) {
	if idChannelOperatorForBilling == "KRUS" {
		var re = regexp.MustCompile(`(\d{2})(\d{2}-.+)`)
		parts := re.FindStringSubmatch(idChannelOperator)
		altIDChannelOperatorForBilling = fmt.Sprintf("RIAS_%s", parts[1])
		return altIDChannelOperatorForBilling
	}
	return "KRUS"
}
