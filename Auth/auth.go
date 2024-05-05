package Auth

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "KombatKode/GolangBsonDB"
  "github.com/google/uuid"
  "reflect"
)

func Auth (w http.ResponseWriter, r *http.Request) {
  token := r.Header.Get("Authorization")
  data := map[string]interface{}{"where": "id", "is": token}
  entryInterface, err := DB.Bson_DB.GetEntries("combatants", data)
  if err != nil {
    return
  }
  entry := reflect.ValueOf(entryInterface)
  bytes, _ := json.Marshal(entry.Interface())
  w.Write(bytes)
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

  entryInterface, err := DB.Bson_DB.CreateEntry("combatants", user)
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
entryInterface, err := DB.Bson_DB.GetEntry("combatants", data)
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
