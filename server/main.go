package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	lua "github.com/yuin/gopher-lua"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

var serverInfo = ServerInfoData{
	Version:  "0.1.0",
	Protocol: 1,
}

var pluginManager PluginManager

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
						Type: "server_info",
						Data: serverInfo,
					})
					pluginManager.Bus.Emit("client_connect", lua.LString(client.ID))
					// c.WriteJSON(Packet{
					// 	Type: "load_scene",
					// 	Data: sceneManager.Scenes["lobby"],
					// })
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

func main() {
	flag.Parse()
	log.SetFlags(0)

	pluginManager = *NewPluginManager()
	defer pluginManager.Close()

	pluginManager.L.SetGlobal("send_packet", pluginManager.L.NewFunction(func(L *lua.LState) int {
		fmt.Println(L.ToString(1)) // this returns nothing for some reason
		lpacket := L.ToTable(2)

		result := Clients.SendPacketToID(
			ClientID(L.ToString(1)),
			Packet{
				Type: lpacket.RawGetString("type").String(),
				Data: L.GetMetatable(lpacket.RawGetString("data")),
			},
		)
		result_int := 0
		if result == true {
			result_int = 1
		}

		return (result_int)
	}))

	if err := pluginManager.LoadPlugins("./plugins"); err != nil {
		panic(err)
	}

	http.HandleFunc("/server", socketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`<h1>Server Shard</h1>`))
