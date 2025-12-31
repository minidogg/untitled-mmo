package main

import (
	"log"
)

type Packet struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
	ID   interface{} `json:"id,omitempty"`
}

func GetField[T any](p *Packet, key string) (T, bool) {
	var zero T

	dataMap, ok := p.Data.(map[string]interface{})
	if !ok {
		log.Printf("Data is not a map for key %s\n", key)
		return zero, false
	}

	rawValue, ok := dataMap[key]
	if !ok {
		log.Printf("Key %s not found\n", key)
		return zero, false
	}

	// Try type assertion
	v, ok := rawValue.(T)
	if ok {
		return v, true
	}

	// Handle JSON numbers (float64) conversion
	if f, isFloat := any(rawValue).(float64); isFloat {
		switch any(zero).(type) {
		case int:
			return any(int(f)).(T), true
		case int64:
			return any(int64(f)).(T), true
		case float32:
			return any(float32(f)).(T), true
		case float64:
			return any(f).(T), true
		}
	}

	log.Printf("Key %s has wrong type: %T\n", key, rawValue)
	return zero, false
}

type ServerInfoData struct {
	Version  string `json:"version"`
	Protocol int    `json:"protocol"`
}

type JoinRejectData struct {
	Version  string `json:"version"`
	Protocol int    `json:"protocol"`
	Message  string `json:"message"`
}

type EntitySnapshot struct {
	ID       EntityID `json:"id"`
	ClientID ClientID `json:"client_id"`
	Input    Input    `json:"input"`

	Position Vec2        `json:"position"`
	Velocity Vec2        `json:"velocity"`
	Size     Vec2        `json:"size"`
	State    EntityState `json:"state"`

	EntityType EntityType  `json:"entity_type"`
	EntityData interface{} `json:"entity_data"`
}

type SnapshotData struct {
	Version    string           `json:"version"`
	Protocol   int              `json:"protocol"`
	ServerTick uint64           `json:"server_tick"`
	SceneID    string           `json:"scene_id"`
	Entities   []EntitySnapshot `json:"entities"`
}
