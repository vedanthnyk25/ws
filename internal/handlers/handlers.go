package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebsocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}

}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type WebsocketConnection struct {
	*websocket.Conn
}

// defines payload sent to the websocket
type WsPayload struct {
	UserName string              `json:"username"`
	Action   string              `json:"action"`
	Message  string              `json:"message"`
	Conn     WebsocketConnection `json:"-"`
}

// defines response sent from the websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// To upgrade http/https connection to a websocket connection
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client connected to endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebsocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

// a goroutine that will run continuously to listen to the websocket and send the payload to the channel
func ListenForWs(conn *WebsocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {

		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

// A goroutine that listens to the channel and sends back json response(broadcast)
func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			//get the list of active users and send back via broadcast
			clients[e.Conn] = e.UserName
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			BroadCastToAll(response)

		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			BroadCastToAll(response)

		case "broadcast":
			response.Action= "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.UserName, e.Message)
			BroadCastToAll(response)	
		}
	}
}

func getUserList() []string {
	var userList []string

	for _, u := range clients {
		if u != "" {
			userList = append(userList, u)
		}
	}

	sort.Strings(userList)
	return userList
}

func BroadCastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("Websocket error")
			_ = client.Close()
			delete(clients, client)
		}
	}
}
