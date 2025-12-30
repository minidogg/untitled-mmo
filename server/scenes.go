package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Entities
type EntityID uint64

type Vec2 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type EntityType int

type EntityState struct {
	Grounded  bool   `json:"grounded"`
	Animation string `json:"animation"`
}

const (
	PlayerEntity EntityType = iota
	EnemyEntity
	PlaceholderEntity
)

type Entity struct {
	ID      EntityID `json:"id"`
	SceneID string   `json:"scene_id"`

	Input  Input   `json:"input"`
	Client *Client `json:"client"`

	Position Vec2        `json:"position"`
	Velocity Vec2        `json:"velocity"`
	State    EntityState `json:"state"`

	EntityType EntityType  `json:"entity_type"`
	EntityData interface{} `json:"entity_data"`

	Dirty  bool
	Remove bool
}

type WorldType int

const (
	Lobby = iota
)

// Rooms & Worlds
type World struct {
	Type  WorldType        `json:"type"`
	Rooms map[string]*Room `json:"rooms"`
}
type Room struct {
	ID       string    `json:"id"`
	Entities []*Entity `json:"entities"`
	TileMap  TileMap   `json:"tile_map"`

	tick    uint64
	running bool
	mu      sync.RWMutex
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

	key := strings.ReplaceAll(file_path, "\\", "/")
	wm.BaseWorlds[key] = world
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

	newWorld := baseWorld
	newWorld.Rooms = make(map[string]*Room)

	for k, room := range baseWorld.Rooms {
		room.running = false
		newWorld.Rooms[k] = room
	}

	wm.ActiveWorlds[id] = newWorld

	for roomID := range newWorld.Rooms {
		room := newWorld.Rooms[roomID]
		go room.Run()
		newWorld.Rooms[roomID] = room
	}

	wm.ActiveWorlds[id] = newWorld
	fmt.Println("World activated")
	return id
}

func NewWorldManager() WorldManager {
	return WorldManager{
		BaseWorlds:   make(map[string]World),
		ActiveWorlds: make(map[string]World),
	}
}

func (wm *WorldManager) CreateEntity(entityData *Entity, worldName, roomName string) {
	world, ok := wm.ActiveWorlds[worldName]
	if !ok {
		fmt.Println("Active world not found: " + worldName)
		return
	}
	room, ok := world.Rooms[roomName]
	if !ok {
		fmt.Println("Room not found: " + roomName)
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	room.Entities = append(room.Entities, entityData)
}

func (r *Room) BroadcastSnapshot(tick uint64) {
	var entitySnapshots []EntitySnapshot
	var activeEntities []*Entity = r.Entities

	for i := range r.Entities {
		e := r.Entities[i]

		if e.Remove {
			activeEntities = slices.Delete(r.Entities, i, i)
			continue
		}

		entitySnapshots = append(entitySnapshots, EntitySnapshot{
			ID:       e.ID,
			ClientID: e.Client.ID,
			Input:    e.Input,

			Position: e.Position,
			Velocity: e.Velocity,
			State:    e.State,

			EntityType: e.EntityType,
			EntityData: e.EntityData,
		})
	}

	r.Entities = activeEntities

	for i := range r.Entities {
		e := r.Entities[i]

		if e.EntityType != PlayerEntity || e.Client.Socket == nil {
			continue
		}

		err := e.Client.Socket.WriteJSON(Packet{
			Type: "room_snapshot",
			Data: SnapshotData{
				Version:    serverInfo.Version,
				Protocol:   serverInfo.Protocol,
				ServerTick: tick,
				SceneID:    r.ID,
				Entities:   entitySnapshots,
			},
		})

		if err != nil {
			fmt.Println("Snapshot to client failed")
		}
	}
}

func (r *Room) StepRoomPhysics() {
	for i := range r.Entities {
		r.Entities[i].StepPhysics()
	}
}

func (r *Room) Run() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.mu.Unlock()

	ticker := time.NewTicker(time.Second / TicksPerSecond)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()

		r.StepRoomPhysics()
		r.BroadcastSnapshot(r.tick)
		r.tick++

		r.mu.Unlock()
	}
}
