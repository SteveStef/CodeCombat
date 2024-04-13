package main
import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"golang.org/x/net/websocket"
)

type Server struct {
  clients []*websocket.Conn
}

func (server *Server) handle_conn(socket *websocket.Conn) {
  fmt.Println("Connection established", socket.RemoteAddr())
  server.clients = append(server.clients, socket)
  server.listen(socket)
}

func (server *Server) listen(socket *websocket.Conn) {
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
    msg := buf[:n]
    fmt.Println("Received message:", string(msg))
    server.broadcast(msg)
    socket.Write([]byte("Thanks for the message"))
  }
}

func (server *Server) broadcast(msg []byte) {
  for _, socket := range server.clients {
    _, err := socket.Write(msg)
    if err != nil {
      fmt.Println(err)
    }
  }
}

func main() {
  fmt.Println("Starting server on port 8080")
  server := Server{ clients: make([]*websocket.Conn, 0), }
  http.Handle("/ws", websocket.Handler(server.handle_conn))

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    template := template.Must(template.ParseFiles("index.html"))
		data := []string{"Hello, World!", "Welcome to our site!", "Enjoy your visit!"}
    template.Execute(w, data)
  })

  http.ListenAndServe(":8080", nil)
}

