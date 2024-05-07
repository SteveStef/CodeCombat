package Compiler

import (
  "os"
  "time"
  "context"
  "os/exec"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "KombatKode/WSServer"
)

func RunCode(w http.ResponseWriter, r *http.Request) {
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var req_struct struct {
    Language string `json:"language"`
    Code string `json:"code"`
  }
  err = json.Unmarshal(body, &req_struct)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  output, wasErr := executeCode(req_struct.Language, req_struct.Code)
  message := map[string]interface{}{"output": string(output), "error": wasErr}

  bytes, _ := json.Marshal(message)
  w.Write(bytes)
}

func executeCode(language string, code string) ([]byte, bool) {
  timeout := WSServer.TIME_LIMIT_RUN * time.Second
  ctx, cancel := context.WithTimeout(context.Background(), timeout)
  defer cancel()

  write_file := "Main." + language
  file, err := os.Create("Execute/" + write_file)
  if err != nil {
    return []byte("Error creating file"), true
  }
  defer file.Close()

  _, err = file.WriteString(code)
  if err != nil {
    return []byte("Error writing to file"), true
  }

  var output []byte
  var cmd *exec.Cmd
  if language == "js" {
    exec_command := "cd Execute && node " + write_file
    cmd = exec.CommandContext(ctx, "bash", "-c", exec_command)
  } else if language == "py" {
    exec_command := "cd Execute && python3 " + write_file
    cmd = exec.CommandContext(ctx, "bash", "-c", exec_command)
  } else if language == "java" {
    exec_command := "cd Execute && javac " + write_file + " && java " + "Main"
    cmd = exec.CommandContext(ctx, "bash", "-c", exec_command)
  } else if language == "cpp" {
    exec_command := "cd Execute && g++ " + write_file + " -o Main && ./Main"
    cmd = exec.CommandContext(ctx, "bash", "-c", exec_command)
  } else {
    return []byte("Error executing code: Invalid language"), true
  }

  output, err = cmd.CombinedOutput()
  os.Remove("Execute/" + write_file)
  if language == "java" {
    os.Remove("Execute/Main.class")
  } else if language == "cpp" {
    os.Remove("Execute/Main")
  }

  if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
      return []byte("Execution timed out"), true
    }
    return output, true
  }
  return output, false
}

