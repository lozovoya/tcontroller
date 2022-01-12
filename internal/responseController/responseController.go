package responseController

import "TController/internal/model"

type Response interface {
	InitReceiversPull(n int)
	AddSource(name, uri string)
	ResponseReceiver(out chan *model.Ticket, id int)
}
