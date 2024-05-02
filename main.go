package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
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

func Auth (w http.ResponseWriter, r *http.Request) {
  cookie := r.Header.Get("Authorization")
  if cookie == "" {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
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
  fmt.Println(uuid.New().String())

  mux.Handle("/ws", CorsMiddleware(websocket.Handler(server.Open_Conn)))
  mux.HandleFunc("/auth", Auth)

  port := os.Getenv("PORT")
  if port == "" {
    log.Fatal("PORT environment variable not set")
  }

  fmt.Println("Listening on port", port)
  log.Fatal(http.ListenAndServe(":" + port, mux))
}



