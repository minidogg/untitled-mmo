package main

import (
	"fmt"
	"sync"
)

// Entities
type EntityID uint64

type Vec2 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type EntityType int

const (
	PlayerEntity EntityType = iota
	EnemyEntity
	PlaceholderEntity
)

type Entity struct {
	ID      EntityID `json:"id"`
	SceneID string   `json:"scene_id"`

	Position Vec2    `json:"position"`
	Velocity Vec2    `json:"velocity"`
	Size     float32 `json:"size"`
	Friction float32 `json:"friction"`
	Static   bool    `json:"static"`

	EntityType EntityType  `json:"entity_type"`
	EntityData interface{} `json:"entity_data"`

	Dirty bool

	mu sync.RWMutex
}

// Rooms & Worlds
type World struct {
	rooms map[string]Room `json:"rooms"`
}
type Room struct {
	ID       string   `json:"id"`
	Type     int      `json:"type"`
	Entities []Entity `json:"entities"`
	TileMap  TileMap  `json:"tile_map"`
}

// Tile Maps
type TileMap map[string]TileMapLayer
type TileMapLayer map[string][]int

// World Manager
type WorldManager struct {
	WorldMap
}

func (*WorldManager) LoadWorld(file_path string) World {
	world, err := ReadJSONFile[World](file_path)
	if err != nil {
		fmt.Println("error loading world ", file_path, err)
	}

	return world
}
