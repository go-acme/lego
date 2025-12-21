package internal

import "encoding/json"

type APIResponse struct {
	Code     string          `json:"code"`
	Message  string          `json:"message"`
	Response json.RawMessage `json:"response"`
}

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClientLogin bool   `json:"client_login"`
}

type ClientIDRequest struct {
	SessionID string `json:"session_id"`
	SysUserID string `json:"sys_userid"`
}

type Zone struct {
	ID        string `json:"id"`
	ServerID  string `json:"server_id"`
	SysUserID string `json:"sys_userid"`
}

type GetTXTRequest struct {
	SessionID string `json:"session_id"`
	PrimaryID struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"primary_id"`
}

type Record struct {
	ID int `json:"id"`
}

type AddTXTRequest struct {
	SessionID    string        `json:"session_id"`
	ClientID     string        `json:"client_id"`
	Params       *RecordParams `json:"params,omitempty"`
	UpdateSerial bool          `json:"update_serial"`
}

type RecordParams struct {
	ServerID string `json:"server_id"`
	Zone     string `json:"zone"`
	Name     string `json:"name"`
	// 'a','aaaa','alias','cname','hinfo','mx','naptr','ns','ds','ptr','rp','srv','txt'
	Type string `json:"type"`
	Data string `json:"data"`
	// "0"
	Aux string `json:"aux"`
	TTL string `json:"ttl"`
	// 'n','y'
	Active string `json:"active"`
	// `2025-12-17 23:35:58`
	Stamp string `json:"stamp"`
	// `1766010947`
	Serial string `json:"serial"`
}

type DeleteTXTRequest struct {
	SessionID    string `json:"session_id"`
	PrimaryID    string `json:"primary_id"`
	UpdateSerial bool   `json:"update_serial"`
}
