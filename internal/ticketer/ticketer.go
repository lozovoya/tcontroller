package ticketer

import (
	"TController/internal/model"
	"context"
)

type Ticket interface {
	CreateTicket(ctx context.Context, ticket *model.Ticket) error
	ReopenTicket(ctx context.Context, ticket *model.Ticket) error
	ChangeTicketStatus(ctx context.Context, ticket *model.Ticket) error
	CheckTicketStatus(ctx context.Context, ticket *model.Ticket) error
	AddNoteToTicket(ctx context.Context, ticket *model.Ticket) error
	CloseTicket(ctx context.Context, ticket *model.Ticket) error
	IDChannelConverter(idChannelOperator string) (string, error)
}
