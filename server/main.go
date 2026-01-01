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
var worldManager = NewWorldManager()
var lobbyMap = ""

func socketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// defer c.Close()

	fmt.Println("new connection!")

	// currently not needed:
	// ip := r.RemoteAddr
	// ip_header := r.Header.Get("cf-connecting-ip")
	// if ip_header != "" {
	// 	ip = ip_header
	// }

	handshaked := false
	client := Clients.GenerateClientFromSocket(c)
	for {
		var msg Packet
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			Clients.RemoveClient(client.ID)
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
						Type: "client_id_assign",
						Data: client.ID,
					})

					c.WriteJSON(Packet{
						Type: "server_info",
						Data: serverInfo,
					})

					var e *Entity = &Entity{
						Client:     client,
						EntityType: PlayerEntity,
						Size: Vec2{
							X: 13.0,
							Y: 21.0,
						},
					}

					client.Entity = e
					worldManager.CreateEntity(e, lobbyMap, "lobby_main")

					c.WriteJSON(Packet{
						Type: "load_room",
						Data: worldManager.ActiveWorlds[lobbyMap].Rooms["lobby_main"],
					})
				} else {
					c.WriteJSON(Packet{
						Type: "join_reject",
						Data: JoinRejectData{
							Protocol: serverInfo.Protocol,
							Version:  serverInfo.Version,
							Message:  "Incorrect protocol version!",
						},
					})
				}
			} else {
				c.WriteJSON(Packet{
					Type: "join_reject",
					Data: JoinRejectData{
						Protocol: serverInfo.Protocol,
						Version:  serverInfo.Version,
						Message:  "No protocol version was received!",
					},
				})
			}

		case "player_tick":
			if client.Entity != nil {
				input, ok := GetField[map[string]interface{}](&msg, "input")

				if !ok {
					fmt.Println("Invalid or nonexistent input object from client")
					break
				}

				client.Entity.Input.Left = input["left"].(float64) != 0
				client.Entity.Input.Right = input["right"].(float64) != 0
				client.Entity.Input.Jump = input["jump"].(float64) != 0
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/server")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	worldManager.LoadWorldMapFolder("./maps")
	lobbyMap = worldManager.CreateNewMap(worldManager.BaseWorlds["maps/lobby.json"])

	http.HandleFunc("/server", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`<h1>Server Shard</h1>`))
