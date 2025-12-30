package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
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
	Grounded bool    `json:"grounded"`

	EntityType EntityType  `json:"entity_type"`
	EntityData interface{} `json:"entity_data"`

	Dirty bool

	mu sync.RWMutex
}

type WorldType int

const (
	Lobby = iota
)

// Rooms & Worlds
type World struct {
	Type  WorldType       `json:"type"`
	Rooms map[string]Room `json:"rooms"`
}
type Room struct {
	ID       string   `json:"id"`
	Entities []Entity `json:"entities"`
	TileMap  TileMap  `json:"tile_map"`
}

// Tile Maps
type TileMap map[string]TileMapLayer
type TileMapLayer map[string][]int

// World Manager
type WorldManager struct {
	BaseWorlds   map[string]World
	ActiveWorlds map[string]World
}

func (wm *WorldManager) LoadWorld(file_path string) World {
	world, err := ReadJSONFile[World](file_path)
	if err != nil {
		fmt.Println("error loading world ", file_path, err)
	}
	wm.BaseWorlds[strings.ReplaceAll(file_path, "\\", "/")] = world

	return world
}

func (wm *WorldManager) LoadWorldMapFolder(rootDir string) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Handle errors that occur while trying to visit a path
			return err
		}

		// Check if it is a regular file and ends with .json
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			fmt.Println("Loading world", path)
			wm.LoadWorld(path)
			fmt.Println("Loaded World", path)
		}

		return nil
	})
}

func (wm *WorldManager) CreateNewMap(baseWorld World) string {
	id := uuid.NewString()
	wm.ActiveWorlds[id] = baseWorld // should create a new clone of the base world i think
	return id
}

func NewWorldManager() WorldManager {
	return WorldManager{
		BaseWorlds:   make(map[string]World),
		ActiveWorlds: make(map[string]World),
	}
}
