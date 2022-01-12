package responseController

import "Smoozzy/TicketSystemController/internal/model"

type Response interface {
	InitReceiversPull(n int)
	AddSource(name, uri string)
	ResponseReceiver(out chan *model.Ticket, id int)
}
