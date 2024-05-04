package main

import (
  "os"
  "log"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "KombatKode/WSServer"
  "github.com/google/uuid"
  "KombatKode/GolangBsonDB"
  "github.com/joho/godotenv"
  "golang.org/x/net/websocket"
)

var ValidDomains = []string{"http://localhost:5173", "http://localhost:8080"}
var db *DB.BsonDB
var wssever *WSServer.Server

func CorsMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    allowed := false

    for _, validOrigin := range ValidDomains {
      if origin == validOrigin {
        allowed = true
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        break
      }
    }

    if !allowed {
      w.WriteHeader(http.StatusForbidden)
      return
    }

    if r.Method == "OPTIONS" {
      w.WriteHeader(http.StatusOK)
      return
    }

    next.ServeHTTP(w, r)
  })
}

func RouteCorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    allowed := false

    for _, validOrigin := range ValidDomains {
      if origin == validOrigin {
        allowed = true
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        break
      }
    }

    if !allowed {
      w.WriteHeader(http.StatusForbidden)
      return
    }

    if r.Method == "OPTIONS" {
      w.WriteHeader(http.StatusOK)
      return
    }

    next(w, r)
  }
}

func Auth (w http.ResponseWriter, r *http.Request) {
  cookie := r.Header.Get("Authorization")
  if cookie == "" {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func Signup(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var request struct {
    Username string `json:"username"`
    Password string `json:"password"`
  }
  err = json.Unmarshal(body, &request)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  uuid := uuid.New().String()

  user := map[string]interface{}{
    "id": uuid, "username": request.Username, "password": request.Password,
    "elo": 0, "language": "", "solved":0, "rank": 0,"level": "Apprentence", "wins": 0, "loses": 0}

  entryInterface, err := db.CreateEntry("combatants", user)
  if err != nil {
    fmt.Println(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  entry, ok := entryInterface.(map[string]interface{})
  if !ok {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  bytes, _ := json.Marshal(entry)
  w.Write(bytes)
}


func Login(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var request struct {
    Username string `json:"username"`
    Password string `json:"password"`
  }
  err = json.Unmarshal(body, &request)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  data := map[string]interface{}{"where": request.Username}
  entryInterface, err := db.GetEntry("combatants", data)
  if err != nil {
    return
  }

  entry, ok := entryInterface.(map[string]interface{})

  if !ok {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  if entry["error"] != nil || entry["password"] != request.Password {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  bytes, _ := json.Marshal(entry)
  w.Write(bytes)
}


func main() {
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Failed to load .env");
  }

  server := WSServer.Server{Clients: []WSServer.User{}, Queue: []WSServer.User{}, Current_Games: []WSServer.Ranked_Set{}}
  db = DB.NewBsonDB(os.Getenv("DATABASE_CONNECTION"))
  //DB.MigrateTable()

  mux := http.NewServeMux()
  mux.Handle("/ws", CorsMiddleware(websocket.Handler(server.Open_Conn)))
  mux.HandleFunc("/auth", Auth)

  mux.HandleFunc("/signup", RouteCorsMiddleware(Signup))
  mux.HandleFunc("/login", RouteCorsMiddleware(Login))

  port := os.Getenv("PORT")
  if port == "" {
    log.Fatal("PORT environment variable not set")
  }

  fmt.Println("Listening on port", port)
  log.Fatal(http.ListenAndServe(":" + port, mux))
}



