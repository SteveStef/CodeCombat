package WSServer

import (
  "fmt"
  "io"
  "encoding/json"
  "golang.org/x/net/websocket"
)

const (
  MAX_PLAYERS = 2

  EASY = 1
  MEDIUM = 2
  HARD = 3

  EASY_TIME = 15
  MEDIUM_TIME = 25
  HARD_TIME = 35

  EASY_POINTS = 10 
  MEDIUM_POINTS = 50
  HARD_POINTS = 100

  TABLE_NAME = "combatants"
  TIME_LIMIT_RUN = 8
  TIME_LIMIT_SUBMIT = 30
)

type Question struct {
  Description string
  Answer string
  Difficulty int
  Category string
  Points int
  Time int
}

type Ranked_Set struct {
  Player1 User
  Player2 User
  Question Question 
}

type Server struct {
  Clients []User
  Queue []User
  Current_Games []Ranked_Set
}

type User struct {
  ID string
  Username string
  Conn *websocket.Conn
}

var Serv *Server

func CreateServer() {
  Serv = &Server{Clients: []User{}, Queue: []User{}, Current_Games: []Ranked_Set{}}
}

func Send(message string, socket *websocket.Conn) error {
  _, err := socket.Write([]byte(message))
  if err != nil {
    return err
  }
  return nil;
}

func (server *Server) Open_Conn(socket *websocket.Conn) {
  fmt.Println("Connection established", socket.RemoteAddr())
  user := User{ ID: fmt.Sprintf("%d", (len(server.Clients) + 1)), Username: "guest", Conn: socket }
  server.Clients = append(server.Clients, user)
  fmt.Println("Number of clients:", server.Clients, len(server.Clients))
  server.Listen(socket)
  return;
}

func (server *Server) Listen(socket *websocket.Conn) {
  buf := make([]byte, 1024)
  for {
    n, err := socket.Read(buf)
    if err != nil {
      if err == io.EOF {
        break
      }
      fmt.Println("Read error:", err)
      continue
    }
    bytes := buf[:n]
    var data map[string]interface{}
    err = json.Unmarshal(bytes, &data)
    if err != nil {
      fmt.Println("Error unmarshalling JSON:", err)
      continue
    }

    if data["type"] == "login" {
      user := User{ ID: data["id"].(string), Username: data["username"].(string), Conn: socket }
      for i, client := range server.Clients {
        if client.Conn == socket {
          server.Clients[i].Username = user.Username
          server.Clients[i].ID = user.ID
          break
        }
      }
    } else if data["type"] == "queue" {
      fmt.Println("Queue request")
    }


    for _, client := range server.Clients {
      fmt.Println(client.Username)
    }

    //server.Handle_Data(data, socket)
    //server.Broadcast(msg)
    //socket.Write([]byte("Thanks for the message"))
  }

  server.Close_Conn(socket)
}

func (server *Server) Close_Conn(socket *websocket.Conn) {
  fmt.Println("Closing connection", socket.RemoteAddr())
  for i, client := range server.Clients {
    if client.Conn == socket {
      server.Clients = append(server.Clients[:i], server.Clients[i+1:]...)
      break
    }
  }
  fmt.Println("Number of clients:", server.Clients, len(server.Clients))
  socket.Close()
}

func (server *Server) Broadcast(msg []byte) {
  for _, user := range server.Clients {
    _, err := user.Conn.Write(msg)
    if err != nil {
      fmt.Println(err)
    }
  }
}

func (server *Server) Remove_from_queue() {
  server.Queue = server.Queue[1:]
}

func (server *Server) Add_to_queue(player User) {
  if len(server.Queue) == 0 {
    fmt.Println("Adding a user to the queue")
    for _, user := range server.Clients {
      if user.Conn == player.Conn {
        server.Queue = append(server.Queue, user)
        break
      }
    }
  } else {
    fmt.Println("Starting a game and broadcasting")

    // pop from queue and start game
    var player2 User
    player2 = server.Queue[0]
    server.Queue = server.Queue[1:]

    // broadcast to both players
    server.BroadCast_Battle("Found a game", player, player2)

  }
}

func (server *Server) BroadCast_Battle(msg string, player1 User, player2 User) {
  _, err := player1.Conn.Write([]byte(msg))
  if err != nil {
    fmt.Println("Unable to connect player 1")
    return
  }
  _, err = player2.Conn.Write([]byte(msg))
  if err != nil {
    fmt.Println("Unable to connect player 2")
    return
  }

  set := Ranked_Set{
    Player1: player1,
    Player2: player2,
    Question: *Get_question(),
  }

  server.Current_Games = append(server.Current_Games, set)
  return 
}

func Get_question() *Question {
  return &Question{
    Description: "What is the capital of France?",
    Answer: "Paris",
    Difficulty: EASY,
    Category: "Geography",
    Points: EASY_POINTS,
    Time: EASY_TIME,
  }
}

