package messageBroker

import (
	"TController/internal/model"
	"context"

	"go.uber.org/zap"
)

const (
	IN_TOPIC = "b2b-TT_IN"
	//IN_TOPIC  = "b2b-TT_OUT"
	OUT_TOPIC = "b2b-TT_OUT"
	Group_ID  = "TicketSystemController"
	USER      = "service_kafkasmz_uk"
	PASS      = "nb$ap#K7dx"
	//USER      = "erkafka"
	//PASS      = "erkafka"
	SCHEMA_ID = 92
)

type Broker interface {
	InitBroker(url string,
		out chan *model.Ticket,
		schemaIN uint32,
		schemaOUT uint32,
		groupID string,
		user string,
		pass string,
		registryURL string,
		lg *zap.Logger) error
	PushMessage(ctx context.Context, topic string, value *model.Ticket) (err error)
	Consumer(ctx context.Context, topic string)
}
