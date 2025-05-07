package main

import (
  "fmt"
  "net/http"
  "github.com/gorilla/websocket"
  "sync"
  "database/sql"
  "encoding/pem"
  "crypto/x509"
  "crypto/rand"
  "crypto/rsa"
  "io"
  _"github.com/go-sql-driver/mysql"
)

var mu sync.RWMutex
var upgrader = websocket.Upgrader{ReadBufferSize: 1024,
                                  WriteBufferSize: 1024,
                                  CheckOrigin: func(r *http.Request) bool { return true }}
var reject_upgrader = websocket.Upgrader{ReadBufferSize: 1024,
                                  WriteBufferSize: 1024,
                                  CheckOrigin: func(r *http.Request) bool { return false }}
var ref_ltr = [52]uint8{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var ref_nb = [10]uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var ref_spechr = [24]uint8{'!', '.', ':', ';', '\\', '-', '%', '*', ',', '_', '/', '<', '>', '=', '[', ']', '\'', '{', '}', '[', ']', '(', ')', '"'}
var banned_char = [2]uint8{'/', ' '}

func GoodId(given_password string) bool {
  var n int = len(given_password)
  if n < 5 {
    return false
  }
  var i int = 0
  var i2 uint
  var cur_val uint8
  var agn bool = true
  for agn && i < n {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 10 && cur_val != ref_nb[i2] {
      i2++
    }
    if i2 < 10 {
      agn = false
    }
    i++
  }
  if agn {
    return false
  }
  agn = true
  i = 0
  for agn && i < n {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 52 && cur_val != ref_ltr[i2] {
      i2++
    }
    if i2 < 52 {
      agn = false
    }
    i++
  }
  i = 0
  if agn {
    return false
  }
  agn = true
  for agn && i < n {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 24 && cur_val != ref_spechr[i2] {
      i2++
    }
    if i2 < 24 {
      agn = false
    }
    i++
  }
  if agn {
    return false
  }
  return true
}

func GoodRoom(given_name string) bool {
  var i2 int
  var cur_val uint8
  for i := 0; i < len(given_name); i++ {
    cur_val = given_name[i]
    for i2 = 0; i2 < 2; i2++ {
      if cur_val == banned_char[i2] {
        return false
      }
    }
  }
  return true 
}

type Clients struct {
  clients map[string]map[*websocket.Conn]bool
}

func HandShakeHandler(db *sql.DB) http.HandlerFunc {
  return func (w http.ResponseWriter, r *http.Request) {
    my_url := r.URL.Path
    var n int = len(my_url)
    if n == 6 {
      w.Write([]byte("Bad URL"))
      return
    }
    if r.Method != "POST" {
      w.Write([]byte("Bad Method"))
      return
    }
    chat_room := ""
    id := ""
    i := 1
    for i < n && my_url[i] != '_' {
      chat_room += string(my_url[i])
      i++
    }
    is_valid := GoodRoom(chat_room)
    if !is_valid {
      w.Write([]byte("Wrong Room"))
      return
    }
    var aes_key string
    var vl_content bool
    content := db.QueryRow("SELECT authorized FROM Status WHERE name=?;", chat_room)
    err := content.Scan(&vl_content)
    if err != nil {
      w.Write([]byte("This room does not exist"))
      return
    }
    i++
    for i < n {
      id += string(my_url[i])
      i++
    }
    is_valid = GoodId(id)
    if !is_valid {
      w.Write([]byte("Wrong Id"))
      return
    }
    var username string
    content = db.QueryRow("SELECT username FROM " + chat_room + " WHERE id=?;", id)
    err = content.Scan(&username)
    if err != nil  {
      w.Write([]byte("Id not existing for this chat room"))
      return
    }
    var vl_content2 bool
    if !vl_content {
      content = db.QueryRow("SELECT already FROM " + chat_room + " WHERE BINARY id=?;", id)
      err = content.Scan(&vl_content2)
      if err != nil {
        fmt.Println(err)
        w.Write([]byte("Something went wrong"))
        return
      }
      if vl_content2 {
        w.Write([]byte("Room not ready yet, retry later"))
        return
      } else {
        raw_pub_key, _, err := r.FormFile("PubKey")
        if err != nil {
          fmt.Println(err)
          w.Write([]byte("Something went wrong"))
          return
        }
        raw_byte_pubKey, err := io.ReadAll(raw_pub_key)
        if err != nil {
          fmt.Println(err)
          w.Write([]byte("Something went wrong"))
          return
        }
        block, _ := pem.Decode(raw_byte_pubKey)
        final_pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
        if err != nil {
          fmt.Println(err)
          w.Write([]byte("Something went wrong"))
          return
        }
        content = db.QueryRow("SELECT aes_key FROM Status WHERE name=?;", chat_room)
        err = content.Scan(&aes_key)
        if err != nil {
          fmt.Println(err)
          w.Write([]byte("Something went wrong"))
          return
        }
        if !is_valid {
          w.Write([]byte("Error during casting"))
          return
        }
        aes_byte := []byte(aes_key)
        ciphered_aes, err := rsa.EncryptPKCS1v15(rand.Reader, final_pubKey, aes_byte)
        if err != nil {
          w.Write([]byte("Something went wrong"))
          return
        }
        _, err = db.Exec("UPDATE " + chat_room + " set already=1 WHERE BINARY id=?;", id)
        if err != nil {
          w.Write([]byte("Something went wrong"))
          return
        }
        content2, err := db.Query("SELECT already FROM " + chat_room + ";")
        if err != nil {
          w.Write([]byte("Something went wrong"))
          return
        }
        for content2.Next() {
          err = content2.Scan(&vl_content)
          if err != nil {
            w.Write([]byte("Something went wrong"))
            return
          }
          if !vl_content {
            break
          }
        }
        if vl_content {
          _, err = db.Exec("UPDATE Status SET authorized=1 WHERE name=?;", chat_room)
          if err != nil {
            w.Write([]byte("Something went wrong"))
            return
          }
          _, err = db.Exec("UPDATE Status SET aes_key=' ' WHERE name=?;", chat_room)
          if err != nil {
            w.Write([]byte("Something went wrong"))
            return
          }
        }
        w.Write(ciphered_aes)
        return
      }
    }
    w.Write([]byte("Ok"))
  }
}

func (clients *Clients) WSHandler(db *sql.DB) http.HandlerFunc {
  return func (w http.ResponseWriter, r *http.Request) {
    my_url := r.URL.Path
    var n int = len(my_url)
    var ws *websocket.Conn
    if n == 6 {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    chat_room := ""
    id := ""
    i := 6
    for i < n && my_url[i] != '_' {
      chat_room += string(my_url[i])
      i++
    }
    is_valid := GoodRoom(chat_room)
    if !is_valid {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    var vl_content bool
    content := db.QueryRow("SELECT authorized FROM Status WHERE name=?;", chat_room)
    err := content.Scan(&vl_content)
    if err != nil {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    i++
    for i < n {
      id += string(my_url[i])
      i++
    }
    is_valid = GoodId(id)
    if !is_valid {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    var username string
    content = db.QueryRow("SELECT username FROM " + chat_room + " WHERE id=?;", id)
    err = content.Scan(&username)
    if err != nil  {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    if !vl_content {
      _, err := reject_upgrader.Upgrade(w, r, nil)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    ws, err = upgrader.Upgrade(w, r, nil)
    if err != nil {
      fmt.Println(err)
      return
    }
    mu.Lock()
    if clients.clients[chat_room] == nil {
      clients.clients[chat_room] = make(map[*websocket.Conn]bool)
    }
    clients.clients[chat_room][ws] = true
    mu.Unlock()
    prefix_msg := username + ": "
    var msg_type int = 1
    var msg []byte
    defer clients.Disconnection(&chat_room, ws, &username)
    for {  
      msg_type, msg, err = ws.ReadMessage()
      if err != nil {
        break
      }
      str_msg := prefix_msg + string(msg)
      msg = []byte(str_msg)
      go clients.Broadcast(&msg, &msg_type, &chat_room)
    }
  }
}

func (clients *Clients) Broadcast(buffr *[]byte, 
                                  msg_type *int,
                                  chat_room *string) {
  var err error
  mu.RLock()
  for cur_ws := range clients.clients[*chat_room] {
    err = cur_ws.WriteMessage(*msg_type, (*buffr))
    if err != nil {
      fmt.Println("error broadcasting message")
      fmt.Println(err)
      return
    }
  }
  mu.RUnlock()
}

func (clients *Clients) Disconnection(chat_room *string, 
                                      ws *websocket.Conn, 
                                      username *string) {
  mu.Lock()
  ws.Close()
  delete(clients.clients[*chat_room], ws)
  mu.Unlock()
}

func ConnectDatabase() (*sql.DB, error) {
  var credentials = "kvv:1234@(localhost:3306)/InFoEnd"
  db, err := sql.Open("mysql", credentials)
  if err != nil {
    return nil, err
  }
  return db, nil
}

func main (){

  db, err := ConnectDatabase();
  if err != nil {
    fmt.Println(err)
    return
  }

  clients := Clients{clients: make(map[string]map[*websocket.Conn]bool)}

  mux := http.NewServeMux()
  mux.HandleFunc("/", HandShakeHandler(db))
  mux.HandleFunc("/chat/", clients.WSHandler(db))
  err = http.ListenAndServe("0.0.0.0:8080", mux)
  if err != nil {
    fmt.Println(err)
    return
  }
}



