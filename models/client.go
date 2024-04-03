package models

import "time"

type Stats struct {
	MaxCount       int64     `json:"max_count"`
	Count          int       `json:"msg_count"`
	FirstMessageAt time.Time `json:"first_message_at"`
	LastMessageAt  time.Time `json:"last_message_at"`
}

type ClientStats struct {
	LastDeliveredId    string `json:"last_delivered_id"`
	LastDeliveredIdIdx int    `json:"last_delivered_id_idx"`
	// number of messages the client is behind the tail
	// by tail we mean a recent message
	CountToTail int `json:"count_to_tail"`
}
