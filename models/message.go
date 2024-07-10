package models

import "encoding/json"

type LogType int

const MessageTypeStdout LogType = 1
const MessageTypeStderr LogType = 2

const MessageTypeInit string = "init"
const MessageTypeLogBulk string = "log_bulk"
const MessageTypeLogSingle string = "log_single"
const MessageTypeClientJoined string = "client_joined"
const MessageTypeClientMsgStatus string = "client_msg_status"

type MessageOrigin struct {
	Port      string `json:"port"`
	File      string `json:"file"`
	ApiSource string `json:"api_source"`
}

type Message struct {
	BaseMessage
	Id          string          `json:"id"`
	Mtype       LogType         `json:"log_type"`
	Content     string          `json:"content"`
	JsonContent json.RawMessage `json:"json_content"`
	IsJson      bool            `json:"is_json"`
	Ts          int64           `json:"ts"`
	Origin      *MessageOrigin  `json:"origin"`
}

type MessageBulk struct {
	BaseMessage
	Messages []Message `json:"messages"`
	Status   Stats     `json:"status"`
}

type ClientJoined struct {
	BaseMessage
	ClientId string `json:"client_id"`
}

type BaseMessage struct {
	MessageType string `json:"message_type"`
}

type ClientMsgStatus struct {
	BaseMessage
	Client ClientStats `json:"client"`
	Stats  Stats       `json:"stats"`
}

type InitMessage struct {
	BaseMessage
	AnalyticsEnabled bool   `json:"analyticsEnabled"`
	AuthRequired     bool   `json:"authRequired"`
	ConfigStr        string `json:"configStr"`
	ApiPrefix        string `json:"apiPrefix"`
}
