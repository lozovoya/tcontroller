package cache

import (
	"Smoozzy/TicketSystemController/internal/model"
)

type CacheRecord struct {
	Source                      string         `json:"source,omitempty"`
	CustomerInternalID          string         `json:"customer_internal_id,omitempty"`
	IDChannelOperatorForBilling string         `json:"tt_for_billing,omitempty"`
	IDChannelOperator           string         `json:"id_channel_operator,omitempty"`
	Description                 string         `json:"description"`
	TTStartTimeTS               int64          `json:"tt_start_time"`
	TTStartTime                 string         `json:"tt_start_time_string"`
	TTClassification            string         `json:"tt_classification"`
	OperatorTTId                string         `json:"tt_number,omitempty"`
	Status                      model.TTStatus `json:"status,omitempty"`
	FileName                    string         `json:"file_name"`
	File                        string         `json:"tt_file,omitempty"`
	Created                     string         `json:"timestamp_start,omitempty"`
	Modified                    string         `json:"timestamp,omitempty"`
}
