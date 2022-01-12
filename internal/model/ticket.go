package model

type RequestType string

const (
	Create RequestType = "create"
	Status RequestType = "status"
	Note   RequestType = "note"
	Wait   RequestType = "wait"
	Reopen RequestType = "reopen"
	Close  RequestType = "close"
	Done   RequestType = "done"
)

type TTStatus string

const (
	Creating TTStatus = "creating" //не заведен ни в одну систему, ожидается ответ
	Error    TTStatus = "error"    //первая система отказала в заведении, запрос направлен во вторую
	Working  TTStatus = "working"
	Waiting  TTStatus = "waiting"
	Closed   TTStatus = "closed"
)

type Ticket struct {
	MessageType                 RequestType `json:"tt_request,omitempty"`
	IDChannelOperatorForBilling string      `json:"tt_for_billing,omitempty"`
	CustomerInternalId          string      `json:"tt_client,omitempty"`
	IDChannelOperator           string      `json:"tt_id_channel_operator,omitempty"`
	Description                 string      `json:"tt_description,omitempty"`
	TTStartTimeTS               int64       `json:"tt_ts_start,omitempty"`
	TTStartTime                 string      `json:"tt_ts_start_string,omitempty"`
	TTClassification            string      `json:"tt_problem_type,omitempty"`
	FileName                    string      `json:"tt_file_name,omitempty"`
	File                        string      `json:"tt_file,omitempty"`
	OperatorTTId                string      `json:"tt_erth,omitempty"`
	EventTimestamp              int64       `json:"tt_ts,omitempty"`
	TimeStampString             string      `json:"tt_ts_string,omitempty"`
	TTStatus                    string      `json:"tt_status,omitempty"`
	Comment                     string      `json:"tt_comment,omitempty"`
	User                        string      `json:"tt_user,omitempty"`
}

type TicketDTO struct {
	Source                      string      `json:"source,omitempty"`
	MessageType                 RequestType `json:"message_type,omitempty"`
	CustomerInternalID          string      `json:"customer_internal_id,omitempty"`
	IDChannelOperatorForBilling string      `json:"tt_for_billing,omitempty"`
	IDChannelOperator           string      `json:"id_channel_operator,omitempty"`
	Description                 string      `json:"description,omitempty"`
	StartTime                   string      `json:"start_time_string,omitempty"`
	StartTimeTS                 int64       `json:"start_time_ts,omitempty"`
	EventTime                   string      `json:"event_time,omitempty"`
	EventTimeTS                 int64       `json:"event_time_timestamp,omitempty"`
	TTClassification            string      `json:"problem_type,omitempty"`
	FileName                    string      `json:"file_name,omitempty"`
	File                        string      `json:"file,omitempty"`
	OperatorTTId                string      `json:"tt_number,omitempty"`
	Status                      string      `json:"status,omitempty"`
	Comment                     string      `json:"comment,omitempty"`
	User                        string      `json:"user,omitempty"`
}
