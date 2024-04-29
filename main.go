package main
import (
	"fmt"
	"io"
	"net/http"
	"golang.org/x/net/websocket"
  "encoding/json"
  "os"
  "github.com/joho/godotenv"
  "github.com/google/uuid"
)

var ValidDomains = []string{"http://localhost:5173", "http://localhost:8080"}

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
  Email string
  Conn *websocket.Conn
}

func CorsMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    domain := r.Header.Get("Origin")
    if domain != "" {
      for _, validDomain := range ValidDomains {
        if domain == validDomain {
          w.Header().Set("Access-Control-Allow-Origin", domain)
          break
        }
      }
    }

    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Allow-Credentials", "true")

    if r.Method == "OPTIONS" {
      return
    }
    next.ServeHTTP(w, r)
  })
}

func Send(message string, socket *websocket.Conn) error {
  _, err := socket.Write([]byte(message))

  if err != nil {
    return err
  }

  return nil;
}

func (server *Server) Open_conn(socket *websocket.Conn) {
  fmt.Println("Connection established", socket.RemoteAddr())
  user := User{ ID: fmt.Sprintf("%d", (len(server.Clients) + 1)), Username: "guest", Email: "", Conn: socket }
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
    fmt.Println("Recieved Data:", data)
    server.Handle_data(data, socket)
    if err != nil {
      fmt.Println("Error unmarshalling JSON:", err)
      continue
    }
    //server.Broadcast(msg)
    //socket.Write([]byte("Thanks for the message"))
  }
  server.Close_conn(socket)
}

func (server *Server) Handle_data(data map[string]interface{}, socket *websocket.Conn) {
  if data["type"] == "login" {
    username := data["username"].(string)
    email := data["email"].(string)
    user := User{ ID: "uuid", Username: username, Email: email, Conn: socket }
    found := false
    for i, client := range server.Clients {
      if client.Conn == socket {
        server.Clients[i] = user
        found = true
        break
      }
    }
    if !found {
      fmt.Println("User not found")
      socket.Write([]byte("User not found"))
    }
  }
}

func (server *Server) Close_conn(socket *websocket.Conn) {
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

func (server *Server) Add_to_queue(socket *websocket.Conn) bool {
  if(len(server.Queue) < MAX_PLAYERS) {
    return true

  } else {
    server.Remove_from_queue()
    return false
  }
}

func Auth (w http.ResponseWriter, r *http.Request) {
  cookie := r.Header.Get("Authorization")

  if cookie == "" {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
  
}

func main() {
  godotenv.Load()

  requiredEnvVars := []string{"PORT"}
  for _, envVar := range requiredEnvVars {
    if os.Getenv(envVar) == "" {
      fmt.Printf("Environment variable %s is not set\n", envVar)
      os.Exit(1)
    }
  }

  server := Server{ Clients: []User{}, Queue: []User{}, Current_Games: []Ranked_Set{} }

  mux := http.NewServeMux()
  http.Handle("/ws", websocket.Handler(server.Open_conn))
  http.HandleFunc("/auth/", Auth)
  handler := CorsMiddleware(mux)

  port := os.Getenv("PORT")
  fmt.Println("Listening on port " + port)
  http.ListenAndServe(":"+port, handler)
}

func get_question() *Question {
  return &Question{
    Description: "What is the capital of France?",
    Answer: "Paris",
    Difficulty: EASY,
    Category: "Geography",
    Points: EASY_POINTS,
    Time: EASY_TIME,
  }
}

