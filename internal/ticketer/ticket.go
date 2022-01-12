package ticketer

import (
	"TController/internal/messageBroker"
	"TController/internal/model"
	"context"
	"fmt"
	"regexp"
)

type ticketWorker struct {
	broker  messageBroker.Broker
	topicIN string
}

func NewTicketWorker(broker messageBroker.Broker, topicIN string) Ticket {
	return &ticketWorker{broker: broker, topicIN: topicIN}
}

func (t *ticketWorker) CreateTicket(ctx context.Context, ticket *model.Ticket) (err error) {

	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("CreateTicket: %w", err)
	}
	return nil
}

func (t *ticketWorker) ReopenTicket(ctx context.Context, ticket *model.Ticket) (err error) {
	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("ReopenTicket: %w", err)
	}
	return nil
}

func (t *ticketWorker) ChangeTicketStatus(ctx context.Context, ticket *model.Ticket) (err error) {
	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("ChangeTicketStatus: %w", err)
	}
	return nil
}

func (t *ticketWorker) CheckTicketStatus(ctx context.Context, ticket *model.Ticket) (err error) {
	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("CheckTicketStatus: %w", err)
	}
	return nil
}

func (t *ticketWorker) AddNoteToTicket(ctx context.Context, ticket *model.Ticket) (err error) {
	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("AddNoteToTicket: %w", err)
	}
	return nil
}

func (t *ticketWorker) CloseTicket(ctx context.Context, ticket *model.Ticket) (err error) {
	err = t.broker.PushMessage(ctx, t.topicIN, ticket)
	if err != nil {
		return fmt.Errorf("CloseTicket: %w", err)
	}
	return nil
}

func (t *ticketWorker) IDChannelConverter(idChannelOperator string) (idChannelOperatorForBilling string, err error) {
	var re = regexp.MustCompile(`(^[a-zA-Z]{3,4})(\d+)-(.+)`)
	parts := re.FindStringSubmatch(idChannelOperator)
	if parts == nil {
		return idChannelOperatorForBilling, fmt.Errorf("IDChannelConverter: wrong IDChannelOperator")
	}
	if (len(parts[1]) == 4) && (len(parts[2]) == 2) {
		idChannelOperatorForBilling = fmt.Sprintf("RIAS_%s", parts[2])
		return idChannelOperatorForBilling, nil
	}
	if (len(parts[1]) == 3) && (len(parts[2]) == 4) {
		idChannelOperatorForBilling = "KRUS"
		return idChannelOperatorForBilling, nil
	}
	return idChannelOperatorForBilling, fmt.Errorf("IDChannelConverter: wrong IDChannelOperator")
}
