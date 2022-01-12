package messageBroker

import (
	"Smoozzy/TicketSystemController/internal/model"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/linkedin/goavro"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"go.uber.org/zap"
)

type kafkaBroker struct {
	conn         kafka.Dialer
	url          string
	reader       kafka.Reader
	writer       kafka.Writer
	out          chan *model.Ticket
	schemaINstr  string
	schemaIN     uint32
	schemaOUTstr string
	schemaOUT    uint32
	groupID      string
	user         string
	pass         string
	registryURL  string
	topicIN      string
	topicOUT     string
	lg           *zap.Logger
}

func NewKafkaBroker() Broker {
	return &kafkaBroker{}
}

func (k *kafkaBroker) InitBroker(url string,
	out chan *model.Ticket,
	schemaIN uint32,
	schemaOUT uint32,
	groupID string,
	user string,
	pass string,
	registryURL string,
	lg *zap.Logger) error {

	k.url = url
	k.out = out
	k.lg = lg
	k.user = user
	k.pass = pass
	k.groupID = groupID

	mechanism := plain.Mechanism{
		Username: k.user,
		Password: k.pass,
	}

	k.conn = kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}
	var err error
	//Импорт из реестра avro схемы для отправки
	k.schemaINstr, err = k.getScheme(registryURL, schemaIN)
	if err != nil {
		return fmt.Errorf("MessageBroker.InitBroker: %w", err)
	}
	//Импорт из реестра avro схемы для приема
	k.schemaOUTstr, err = k.getScheme(registryURL, schemaOUT)
	if err != nil {
		return fmt.Errorf("MessageBroker.InitBroker: %w", err)
	}
	return nil
}

func (k *kafkaBroker) PushMessage(ctx context.Context, topic string, ticket *model.Ticket) (err error) {
	mechanism := plain.Mechanism{
		Username: k.user,
		Password: k.pass,
	}

	writer := kafka.Writer{
		Addr:         kafka.TCP(k.url),
		Topic:        topic,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxAttempts:  3,
		RequiredAcks: kafka.RequireAll,
		Transport: &kafka.Transport{
			SASL: mechanism,
		},
	}
	defer writer.Close()
	log.Printf("send message: %v", ticket)
	message, err := k.ticketToBinaryConverter(ticket)
	if err != nil {
		return fmt.Errorf("messageBroker.PushMessage: %w", err)
	}
	err = writer.WriteMessages(ctx, kafka.Message{Value: message})
	if err != nil {
		return fmt.Errorf("messageBroker.PushMessage: %w", err)
	}
	return nil
}

func (k *kafkaBroker) Consumer(ctx context.Context, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{k.url},
		Topic:       topic,
		Partition:   0,
		GroupID:     k.groupID,
		StartOffset: -1,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		Dialer:      &k.conn,
	})
	go func() {
		for {
			var ticket *model.Ticket
			message, err := reader.ReadMessage(ctx)
			k.lg.Info("got message from kafka")
			if err != nil {
				k.lg.Error("Consumer.ReadMessage", zap.Error(err)) //todo сделать канал для приема ошибок из kafka
				continue
			}

			ticket, err = k.binaryToTicketConverter(message.Value)
			if err != nil {
				k.lg.Error("Consumer.ReadMessage", zap.Error(err))
				continue
			}
			k.out <- ticket
		}
	}()
}

func (k *kafkaBroker) decodeAvro(bytes []byte) (interface{}, error) {
	avroCodec, err := goavro.NewCodec(k.schemaINstr)
	if err != nil {
		return bytes, fmt.Errorf("messageBroker.Decode: %w", err)
	}
	bytes = bytes[5:]
	data, _, err := avroCodec.NativeFromBinary(bytes)
	if err != nil {
		return bytes, fmt.Errorf("messageBroker.Decode: %w", err)
	}
	return data, nil
}

func (k *kafkaBroker) encodeAvro(data map[string]interface{}) ([]byte, error) {
	bytes := make([]byte, 5)
	avroCodec, err := goavro.NewCodec(k.schemaINstr)
	if err != nil {
		return bytes, fmt.Errorf("messageBroker.PushMessage: %w", err)
	}
	log.Printf("Created schema %s", avroCodec.Schema())
	bytes[0] = 0
	binary.BigEndian.PutUint32(bytes[1:5], k.schemaIN)
	message, err := avroCodec.BinaryFromNative(bytes, data)
	if err != nil {
		return bytes, fmt.Errorf("messageBroker.PushMessage: %w", err)
	}
	return message, nil
}

func (k *kafkaBroker) binaryToTicketConverter(bytes []byte) (*model.Ticket, error) {
	var ticket *model.Ticket
	decodedMessage, err := k.decodeAvro(bytes)
	if err != nil {
		return ticket, fmt.Errorf("binaryToTicketConverter: %w", err)
	}
	unwrapMessageStage1 := decodedMessage.(map[string]interface{})
	m := make(map[string]string)
	for key, value := range unwrapMessageStage1 {
		if value == nil {
			continue
		}
		unwrapMessageStage2 := value.(map[string]interface{})
		if unwrapMessageStage2["string"] == nil {
			continue
		}
		str := unwrapMessageStage2["string"].(string)
		m[key] = str
	}

	ttStartTimeTS64, err := strconv.ParseInt(m["tt_ts_start"], 10, 64)
	eventTimestamp64, err := strconv.ParseInt(m["tt_ts"], 10, 64)

	ticket = &model.Ticket{
		IDChannelOperatorForBilling: m["tt_for_billing"],
		CustomerInternalId:          m["tt_client"],
		IDChannelOperator:           m["tt_id_channel_operator"],
		Description:                 m["tt_description"],
		TTStartTimeTS:               ttStartTimeTS64,
		TTClassification:            m["tt_problem_type"],
		FileName:                    m["tt_file_name"],
		File:                        m["tt_file"],
		OperatorTTId:                m["tt_erth"],
		EventTimestamp:              eventTimestamp64,
		TTStatus:                    m["tt_status"],
		Comment:                     m["tt_comment"],
		User:                        m["tt_user"],
	}
	switch m["tt_request"] {
	case "create":
		ticket.MessageType = model.Create
	case "close":
		ticket.MessageType = model.Close
	case "status":
		ticket.MessageType = model.Status
	case "Reopen":
		ticket.MessageType = model.Reopen
	case "wait":
		ticket.MessageType = model.Wait
	case "note":
		ticket.MessageType = model.Note
	default:
		ticket.MessageType = model.Note
	}
	return ticket, nil
}

func (k *kafkaBroker) ticketToBinaryConverter(ticket *model.Ticket) ([]byte, error) {
	var requestType string
	requestType = string(ticket.MessageType)
	m := map[string]interface{}{
		"tt_request": map[string]interface{}{
			"string": requestType,
		},
		"tt_for_billing": map[string]interface{}{
			"string": ticket.IDChannelOperatorForBilling,
		},
		"tt_client": map[string]interface{}{
			"string": ticket.CustomerInternalId,
		},
		"tt_id_channel_operator": map[string]interface{}{
			"string": ticket.IDChannelOperator,
		},
		"tt_description": map[string]interface{}{
			"string": ticket.Description,
		},
		"tt_ts_start": map[string]interface{}{
			"long": ticket.TTStartTimeTS,
		},
		"date_in_string": map[string]interface{}{
			"string": ticket.TTStartTime,
		},
		"tt_problem_type": map[string]interface{}{
			"string": ticket.TTClassification,
		},
		"tt_file_name": map[string]interface{}{
			"string": ticket.FileName,
		},
		"tt_file": map[string]interface{}{
			"string": ticket.File,
		},
		"tt_erth": map[string]interface{}{
			"string": ticket.OperatorTTId,
		},
		"tt_ts": map[string]interface{}{
			"long": ticket.EventTimestamp,
		},
		"tt_status": map[string]interface{}{
			"string": ticket.TTStatus,
		},
		"tt_comment": map[string]interface{}{
			"string": ticket.Comment,
		},
		"tt_user": map[string]interface{}{
			"string": ticket.User,
		},
	}
	message, err := k.encodeAvro(m)
	if err != nil {
		return nil, fmt.Errorf("messageBroker.ticketToBinaryConverter: %w", err)
	}
	return message, nil
}

func (k *kafkaBroker) getScheme(registryURL string, schemeID uint32) (string, error) {
	reqURL := fmt.Sprintf("%s%d", registryURL, schemeID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("MessageBroker.getScheme: %w", err)
	}
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*15)
	defer cancel()
	req = req.WithContext(ctx)
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("MessageBroker.getScheme: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("MessageBroker.getScheme: %w", err)
	}
	scheme := struct {
		Scheme string `json:"schema"`
	}{}
	err = json.Unmarshal(body, &scheme)
	if err != nil {
		return "", fmt.Errorf("MessageBroker.getScheme: %w", err)
	}
	log.Printf("Imported scheme with ID %d,  scheme: %+v", schemeID, scheme)
	return scheme.Scheme, nil
}
