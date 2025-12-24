package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func WriteJSONFile[T any](filename string, data T) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func ReadJSONFile[T any](filename string) (T, error) {
	var data T

	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

type SceneMainFile struct {
	TopLeft Vec2   `json:"id"`
	TileMap string `json:"tilemap"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

func LoadScene(filePath string) *Scene {

	metadata, err := ReadJSONFile[SceneMainFile](filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	tilemap, err := ReadTileMap(metadata.TileMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	scene_type := SceneLobby
	if metadata.Type == "Dungeon" {
		scene_type = SceneDungeon
	}
	scene := sceneManager.CreateScene(metadata.Name, scene_type, tilemap)

	log.Printf("Created scene with ID: %s (size: %dx%d)", scene.ID, tilemap.Width, tilemap.Height)

	// staticEntity := CreateStaticPlaceholderEntity(1238, scene.ID, 0, 0)
	// scene.AddEntity(staticEntity)

	return scene
}
