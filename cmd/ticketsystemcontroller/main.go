package main

import (
	"Smoozzy/TicketSystemController/internal/api/httpserver"
	v1 "Smoozzy/TicketSystemController/internal/api/httpserver/v1"
	timer2 "Smoozzy/TicketSystemController/internal/timer"
	"time"

	"Smoozzy/TicketSystemController/internal/cache"
	"Smoozzy/TicketSystemController/internal/messageBroker"
	"Smoozzy/TicketSystemController/internal/model"
	"Smoozzy/TicketSystemController/internal/responseController"
	"Smoozzy/TicketSystemController/internal/ticketer"
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Params = struct {
	//General
	Port string `env:"TTS_PORT" envDefault:"14801"`
	Host string `env:"TTS_HOST" envDefault:"0.0.0.0"`

	//Kafka
	BrokerURL string `env:"BROKER_URL" envDefault:"10.101.15.110:9094"`
	//BrokerURL string `env:"BROKER_URL" envDefault:"esb-3.ertelecom.ru:9094"`
	OutTopic string `env:"OUT_TOPIC" envDefault:"b2b-TT_OUT"`
	InTopic  string `env:"IN_TOPIC" envDefault:"b2b-TT_IN"`
	//InTopic         string `env:"IN_TOPIC" envDefault:"b2b-TT_OUT" //для тестирования ответо`
	RegistryURL     string `env:"REGISTRY_URL" envDefault:"http://10.101.15.110:8081/schemas/ids/"`
	BrokerUser      string `env:"BROKER_USER" envDefault:"service_kafkasmz_uk"`
	BrokerPass      string `env:"BROKER_PASS" envDefault:"nb$ap#K7dx"`
	InSchemeID      int    `env:"IN_SCHEME" envDefault:"92"`
	OutSchemeID     int    `env:"OUT_SCHEME" envDefault:"71"`
	BrokerGroupID   string `env:"BROKER_GROUP" envDefault:"TicketSystemController"`
	ConsumerStreams int    `env:"CONSUMER_STREAMS" envDefault:"5"`

	//Redis
	CacheDSN string `env:"CACHE_DSN" envDefault:"redis://HQFQBb6fDcUK@dev-redis-master/0"`
	CacheTTL int64  `env:"CACHE_TTL" envDefault:"259200"` //3 дня

	//SberAPI
	SberAPIID  string `env:"SBER_API_URI" envDefault:"sberapi"`
	SberAPIURI string `env:"SBER_API_URI" envDefault:"http://smoozzy-sberapi-service:14800"`
}

func main() {

	var controllerParameters Params
	err := env.Parse(&controllerParameters)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err = execute(controllerParameters); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func execute(controllerParameters Params) (err error) {
	lg := zap.NewExample()
	defer lg.Sync()

	cachePool := cache.InitCache(controllerParameters.CacheDSN)
	cache := cache.NewRedisCache(cachePool, controllerParameters.CacheTTL, lg)
	cacheController := v1.NewCacheController(cache, lg)

	out := make(chan *model.Ticket)
	broker := messageBroker.NewKafkaBroker()
	broker.InitBroker(controllerParameters.BrokerURL,
		out,
		uint32(controllerParameters.InSchemeID),
		uint32(controllerParameters.OutSchemeID),
		controllerParameters.BrokerGroupID,
		controllerParameters.BrokerUser,
		controllerParameters.BrokerPass,
		controllerParameters.RegistryURL, lg)

	go func() {
		broker.Consumer(context.Background(), controllerParameters.OutTopic)
	}()

	//ctx := context.Background()
	ctx := context.TODO()
	timer := timer2.NewTimer(ctx, time.Minute*2, cache)
	go func() {
		for {
			timer.FindExpired()
			log.Println("Cache checked for expired records")
			time.Sleep(time.Minute)
		}
	}()

	ticketWorker := ticketer.NewTicketWorker(broker, controllerParameters.InTopic)
	ticketController := v1.NewTicketer(ticketWorker, cache, lg)

	receiver := responseController.NewReceiver(out, cache, ticketWorker, lg)
	receiver.InitReceiversPull(controllerParameters.ConsumerStreams)
	receiver.AddSource(controllerParameters.SberAPIID, controllerParameters.SberAPIURI)

	router := httpserver.NewRouter(chi.NewRouter(), lg, ticketController, cacheController)
	server := http.Server{
		Addr:        net.JoinHostPort(controllerParameters.Host, controllerParameters.Port),
		Handler:     &router,
		IdleTimeout: time.Second * 30,
	}

	return server.ListenAndServe()
}
