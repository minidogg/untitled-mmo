package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var serverInfo = ServerInfoData{
	Version:  "0.1.0",
	Protocol: 1,
}

// Global scene manager
var sceneManager *SceneManager

func socketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	fmt.Println("new connection!")

	// currently not needed:
	// ip := r.RemoteAddr
	// ip_header := r.Header.Get("cf-connecting-ip")
	// if ip_header != "" {
	// 	ip = ip_header
	// }

	handshaked := false
	for {
		var msg Packet
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}
		fmt.Println(msg)

		switch msg.Type {

		case "hello":
			if handshaked == true {
				break
			}
			handshaked = true
			protocol, ok := GetField[int](&msg, "protocol")
			if ok {
				if protocol == serverInfo.Protocol {
					c.WriteJSON(Packet{
						Type: "server_info",
						Data: serverInfo,
					})
					c.WriteJSON(Packet{
						Type: "load_scene",
						Data: sceneManager.Scenes["lobby"],
					})
				} else {
					c.WriteJSON(Packet{
						Type: "join_reject",
						Data: JoinRejectData{
							ProtocolVersion: serverInfo.Protocol,
							Version:         serverInfo.Version,
							Message:         "Incorrect protocol version!",
						},
					})
				}
			} else {
				c.WriteJSON(Packet{
					Type: "join_reject",
					Data: JoinRejectData{
						ProtocolVersion: serverInfo.Protocol,
						Version:         serverInfo.Version,
						Message:         "No protocol version was received!",
					},
				})
			}

		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/server")
}

// currently ai generated, i aint hand writing a level in hard code
func initializeScenes() {
	sceneManager = &SceneManager{
		Scenes: make(map[string]*Scene),
	}

	// Create lobby scene - simple flat ground with platforms
	lobbyTilemap := &TileMap{
		Width:  100,
		Height: 30,
		Tiles:  make([]TileID, 100*30),
	}
	// Fill with air (tile ID 0)
	for i := range lobbyTilemap.Tiles {
		lobbyTilemap.Tiles[i] = 1
	}
	// Create ground at the bottom (last 3 rows)
	for y := 27; y < 30; y++ {
		for x := 0; x < 100; x++ {
			lobbyTilemap.Tiles[y*100+x] = 2 // ground tile
		}
	}
	// Add some floating platforms
	for x := 20; x < 30; x++ {
		lobbyTilemap.Tiles[20*100+x] = 2
	}
	for x := 50; x < 65; x++ {
		lobbyTilemap.Tiles[15*100+x] = 2
	}
	for x := 75; x < 85; x++ {
		lobbyTilemap.Tiles[18*100+x] = 2
	}

	lobbyScene := sceneManager.CreateScene("lobby", SceneLobby, lobbyTilemap)
	log.Printf("Created lobby scene with ID: %s (size: %dx%d)", lobbyScene.ID, lobbyTilemap.Width, lobbyTilemap.Height)

	staticEntity := CreateStaticPlaceholderEntity(1238, "lobby", 0, 0)
	lobbyScene.AddEntity(staticEntity)

	// Create a dungeon scene - side-scrolling level with walls, floors, and platforms
	dungeonTilemap := &TileMap{
		Width:  200,
		Height: 40,
		Tiles:  make([]TileID, 200*40),
	}
	// Fill with air (tile ID 0)
	for i := range dungeonTilemap.Tiles {
		dungeonTilemap.Tiles[i] = 1
	}

	// Create floor at bottom (last 5 rows are solid ground)
	for y := 35; y < 40; y++ {
		for x := 0; x < 200; x++ {
			dungeonTilemap.Tiles[y*200+x] = 3 // dungeon floor
		}
	}

	// Add ceiling at top (first 2 rows)
	for y := 0; y < 2; y++ {
		for x := 0; x < 200; x++ {
			dungeonTilemap.Tiles[y*200+x] = 3 // dungeon ceiling
		}
	}

	// Add left and right walls
	for y := 0; y < 40; y++ {
		dungeonTilemap.Tiles[y*200+0] = 3   // left wall
		dungeonTilemap.Tiles[y*200+199] = 3 // right wall
	}

	// Add some platforms throughout the level
	platformData := []struct{ x, y, length int }{
		{15, 28, 10},  // low platform
		{30, 22, 12},  // mid platform
		{50, 18, 8},   // high platform
		{65, 25, 15},  // long mid platform
		{90, 20, 10},  // platform
		{110, 15, 8},  // high platform
		{130, 28, 12}, // low platform
		{150, 22, 10}, // mid platform
		{170, 18, 15}, // high platform
	}

	for _, plat := range platformData {
		for x := plat.x; x < plat.x+plat.length && x < 200; x++ {
			dungeonTilemap.Tiles[plat.y*200+x] = 4 // platform tile
		}
	}

	// Add some pillars/obstacles
	for y := 30; y < 35; y++ {
		dungeonTilemap.Tiles[y*200+40] = 3  // pillar 1
		dungeonTilemap.Tiles[y*200+120] = 3 // pillar 2
	}

	dungeonScene := sceneManager.CreateScene("dungeon_01", SceneDungeon, dungeonTilemap)
	log.Printf("Created dungeon scene with ID: %s (size: %dx%d)", dungeonScene.ID, dungeonTilemap.Width, dungeonTilemap.Height)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	initializeScenes()

	http.HandleFunc("/server", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`<h1>Server Shard</h1>`))
