package main

import (
  "os"
  "log"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "KombatKode/Auth"
  "KombatKode/Compiler"
  "KombatKode/WSServer"
  "KombatKode/GolangBsonDB"
  "github.com/joho/godotenv"
  "golang.org/x/net/websocket"
)

var ValidDomains = []string{"http://localhost:5173", "http://localhost:8080"}

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

func RankedGame(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {  
    w.WriteHeader(http.StatusBadRequest)
    return;
  }
  var reqstruct struct {
    Username string `json:"username"`
  }
  err = json.Unmarshal(body, &reqstruct)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return;
  }
  //wssever.Add_to_queue()
  fmt.Println("Ranked Game Requested by", reqstruct.Username)
  w.WriteHeader(http.StatusOK)
}

func main() {
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Failed to load .env");
  }

  WSServer.CreateServer()
  DB.Bson_DB = DB.NewBsonDB(os.Getenv("DATABASE_CONNECTION"))
  //DB.MigrateTable()

  mux := http.NewServeMux()

  mux.Handle("/ws", CorsMiddleware(websocket.Handler(WSServer.Serv.Open_Conn)))

  mux.HandleFunc("/auth", RouteCorsMiddleware(Auth.Auth))
  mux.HandleFunc("/signup", RouteCorsMiddleware(Auth.Signup))
  mux.HandleFunc("/login", RouteCorsMiddleware(Auth.Login))
  mux.HandleFunc("/run", RouteCorsMiddleware(Compiler.RunCode))
  mux.HandleFunc("/ranked", RouteCorsMiddleware(RankedGame))

  port := os.Getenv("PORT")
  if port == "" {
    log.Fatal("PORT environment variable not set")
  }

  fmt.Println("Listening on port", port)
  log.Fatal(http.ListenAndServe(":" + port, mux))
}



