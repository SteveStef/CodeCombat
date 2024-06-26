package Auth

import (
  "fmt"
  "time"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "github.com/google/uuid"
  "KombatKode/GolangBsonDB"
  "KombatKode/WSServer"
)

func Auth (w http.ResponseWriter, r *http.Request) {
  token := r.Header.Get("Authorization")
  data := map[string]interface{}{"where": "id", "is": token}
  entrys, err := DB.Bson_DB.GetEntries("combatants", data)

  if err != nil { 
    fmt.Println("working")
    w.WriteHeader(http.StatusInternalServerError)
    return 
  }

  if len(entrys) == 0 {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  currentGame, found := WSServer.Serv.CheckCurrentGames(entrys[0]["username"].(string))
  currentTime := int(time.Now().Unix())

  timeLeft := currentGame.Question.Time * 60 - (currentTime - currentGame.CurrentTime)

  if found {
    data = map[string]interface{}{"player": entrys[0], "game": currentGame, "timeLeft": timeLeft}
    bytes, _ := json.Marshal(data)
    w.Write(bytes)
  } else {
    bytes, _ := json.Marshal(entrys[0])
    w.Write(bytes)
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

  aLog := make([]int8, 32)

  user := map[string]interface{}{
    "id": uuid, "username": request.Username, "password": request.Password,
    "elo": 0, "language": "", "solved":0, "rank": 0,"level": "Apprentence", "wins": 0, "loses": 0,
    "eloHistory": []int32{}, "activityLog": aLog,}

  entry, err := DB.Bson_DB.CreateEntry("combatants", user)
  if err != nil {
    fmt.Println(err)
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
  entry, err := DB.Bson_DB.GetEntry("combatants", data)
  if err != nil {
    return
  }

  if entry["error"] != nil || entry["password"] != request.Password {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  bytes, _ := json.Marshal(entry)
  w.Write(bytes)
}

