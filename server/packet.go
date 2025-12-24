package main

type Packet struct {
	Cmd  string      `json:"cmd"`
	Name interface{} `json:"name,omitempty"`
	Val  interface{} `json:"val,omitempty"`
	ID   interface{} `json:"id,omitempty"`
}
