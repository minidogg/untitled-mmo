package main

import "sync"

// Network owner ship
type OwnershipType int

const (
	OwnerServer OwnershipType = iota
	OwnerClient
)

type NetworkOwner struct {
	OwnerType OwnershipType
	ClientID  string
}

// Entities
type EntityID uint64

type Vec2 struct {
	X, Y float32
}

type EntityType int

const (
	PlayerEntity EntityType = iota
	EnemyEntity
)

type Entity struct {
	ID      EntityID
	SceneID string

	Position Vec2
	Velocity Vec2
	Friction float32
	Static   bool

	EntityType EntityType
	EntityData interface{}

	Owner NetworkOwner
	Dirty bool

	mu sync.RWMutex
}

func (entity *Entity) SetOwner(owner NetworkOwner) {
	entity.mu.Lock()
	defer entity.mu.Unlock()
	entity.Owner = owner
	entity.Dirty = true
}

// Tiles and tilemaps
type TileID uint16

type TileMap struct {
	Width  int
	Height int
	Tiles  []TileID // row-major (y*Width + x)
	mu     sync.RWMutex
}

func (tm *TileMap) Get(x, y int) TileID {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.Tiles[y*tm.Width+x]
}

func (tm *TileMap) Set(x, y int, tile TileID) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Tiles[y*tm.Width+x] = tile
}

// Scenes
type SceneType int

const (
	SceneLobby SceneType = iota
	SceneDungeon
)

type Scene struct {
	ID       string
	Type     SceneType
	TileMap  *TileMap
	Entities map[EntityID]*Entity
	mu       sync.RWMutex
}

func (scene *Scene) GetEntity(id EntityID) (*Entity, bool) {
	scene.mu.RLock()
	defer scene.mu.RUnlock()

	e, ok := scene.Entities[id]
	return e, ok
}

func (scene *Scene) AddEntity(e *Entity) {
	scene.mu.Lock()
	defer scene.mu.Unlock()

	scene.Entities[e.ID] = e
}

func (scene *Scene) RemoveEntity(id EntityID) {
	scene.mu.Lock()
	defer scene.mu.Unlock()
	entity := scene.Entities[id]
	entity.mu.Lock()
	delete(scene.Entities, id)
	entity.mu.Unlock()
}

// TODO: make this only snapshot areas in a radius around clients
func (s *Scene) Snapshot() []Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Entity, 0, len(s.Entities))
	for _, e := range s.Entities {
		out = append(out, *e) // copy
	}
	return out
}

func (s *Scene) Tick(dt float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.Entities {
		if e.Owner.OwnerType == OwnerServer {
			moveX := e.Velocity.X / e.Friction * dt
			moveY := e.Velocity.Y / e.Friction * dt

			e.Position.X += moveX
			e.Position.Y += moveY

			e.Velocity.X -= moveX
			e.Velocity.Y -= moveY

			e.Dirty = true
		}
	}
}

// Scene manager
type SceneManager struct {
	Scenes map[string]*Scene
	mu     sync.RWMutex
}

func (sceneManager *SceneManager) CreateScene(sceneId string, sceneType SceneType, tileMap *TileMap) *Scene {
	scene := &Scene{
		ID:       sceneId,
		Type:     sceneType,
		TileMap:  tileMap,
		Entities: make(map[EntityID]*Entity),
	}
	sceneManager.mu.Lock()
	sceneManager.Scenes[sceneId] = scene
	sceneManager.mu.Unlock()
	return scene
}
