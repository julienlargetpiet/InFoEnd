package main

import (
  "fmt"
  "github.com/gdamore/tcell/v2"
  "github.com/rivo/tview"
  "crypto/aes"
  "crypto/cipher"
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "mime/multipart"
  "bytes"
  "encoding/pem"
  "net/http"
  "github.com/gorilla/websocket"
  "os"
  "io"
  "time"
  //"unicode/utf8"
)

var global_aes_key string = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
var max_row int = 17
var delta_max_row int = max_row - 1
var ref_ltr = [52]uint8{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var ref_nb = [10]uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var ref_spechr = [23]uint8{'.', ':', ';', '\\', '-', '%', '*', ',', '_', '/', '<', '>', '=', '[', ']', '\'', '{', '}', '[', ']', '(', ')', '"'}
var banned_char = [2]uint8{'/', ' '}
var ref_err_resp = [9]string{"Something went wrong", 
                            "Error during casting", 
                            "Room not ready yet, retry later", 
                            "Id not existing for this chat room", 
                            "Wrong Id",
                            "This room does not exist",
                            "Wrong Room",
                            "Bad Method",
                            "Bad URL"}

func StringToInt32(x string) int32 {
  var ref_nb = [10]uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
  var rtn_val int32 = 0
  var lngth int = len(x)
  var i2 int32
  var cur_rn uint8
  var i int
  for i = 0; i + 1 < lngth; i++ {
    cur_rn = x[i]
    i2 = 0
    for cur_rn != ref_nb[i2]{
      i2++
    }
    rtn_val += i2
    rtn_val *= 10
  }
  cur_rn = x[i]
  i2 = 0
  for cur_rn != ref_nb[i2]{
    i2++
  }
  rtn_val += i2
  return rtn_val
}

func URLToUTF8(x string) string {
  cur_vl := ""
  var int_vl int32
  var rtn_rune []rune
  for i:= 0; i < len(x); i++ {
    if x[i] != '-' {
      cur_vl += string(x[i])
    } else {
      int_vl = StringToInt32(cur_vl)
      rtn_rune = append(rtn_rune, rune(int_vl))
      cur_vl = ""
    }
  }
  int_vl = StringToInt32(cur_vl)
  rtn_rune = append(rtn_rune, rune(int_vl))
  fmt.Println("end", rtn_rune)
  return string(rtn_rune)
}

func cipherer(x *string, secret_key *string) string {

  cur_aes, err := aes.NewCipher([]byte(*secret_key))
  if err != nil {
    fmt.Println(err)
    return ""
  }
  cur_gcm, err := cipher.NewGCM(cur_aes)
  if err != nil {
    fmt.Println(err)
    return ""
  }
  cur_nonce := make([]byte, cur_gcm.NonceSize())
  _, err = rand.Read(cur_nonce)
  if err != nil {
    fmt.Println(err)
    return ""
  }

  cipher_data := cur_gcm.Seal(cur_nonce, cur_nonce, []byte(*x), nil)

  //fmt.Println("OO", cipher_data)

  return string(cipher_data)
}

func decipherer(x *string, secret_key *string) string {
  cur_aes, err := aes.NewCipher([]byte(*secret_key))
  if err != nil {
    fmt.Println(err)
    return ""
  }
  cur_gcm, err := cipher.NewGCM(cur_aes)
  if err != nil {
    fmt.Println(err)
    return ""
  }
  nonce_size := cur_gcm.NonceSize()
  cur_nonce := (*x)[:nonce_size]
  cipher_data := (*x)[nonce_size:]
  
  deciphered_data, err := cur_gcm.Open(nil, []byte(cur_nonce), []byte(cipher_data), nil)
  if err != nil {
    fmt.Println(err)
    return ""
  }
  return string(deciphered_data)
}

func ByteCipherer(x *string, secret_key *string) []byte {

  cur_aes, err := aes.NewCipher([]byte(*secret_key))
  if err != nil {
    fmt.Println(err)
    return []byte{}
  }
  cur_gcm, err := cipher.NewGCM(cur_aes)
  if err != nil {
    fmt.Println(err)
    return []byte{}
  }
  cur_nonce := make([]byte, cur_gcm.NonceSize())
  _, err = rand.Read(cur_nonce)
  if err != nil {
    fmt.Println(err)
    return []byte{}
  }

  cipher_data := cur_gcm.Seal(cur_nonce, cur_nonce, []byte(*x), nil)

  return cipher_data
}

func ByteDecipherer(x *[]byte, secret_key *string) string {
  cur_aes, err := aes.NewCipher([]byte(*secret_key))
  if err != nil {
    fmt.Println(err)
    return ""
  }
  cur_gcm, err := cipher.NewGCM(cur_aes)
  if err != nil {
    fmt.Println(err)
    return ""
  }
  nonce_size := cur_gcm.NonceSize()
  cur_nonce := (*x)[:nonce_size]
  cipher_data := (*x)[nonce_size:]
  
  deciphered_data, err := cur_gcm.Open(nil, []byte(cur_nonce), cipher_data, nil)
  if err != nil {
    fmt.Println(err)
    return ""
  }
  return string(deciphered_data)
}

func GoodPort(x string) bool {
  var i2 int
  for i := 0; i < len(x); i++ {
    i2 = 0
    for i2 < 10 {
      if x[i] != ref_nb[i2] {
        i2++
      } else {
        break
      }
    }
    if i2 == 10 {
      return false
    }
  }
  int_port := StringToInt32(x)
  if int_port < 5000 || int_port > 90000 {
    return false
  }
  return true
}

func GoodIP(x *string) bool {
  var n int  = len(*x)
  var i int = 0
  var i2 int
  var cur_val string
  for I := 0; I < 3; I++ {
    cur_val = ""
    for i < n && (*x)[i] != '.' {
      i2 = 0
      for i2 < 10 {
        if ref_nb[i2] != (*x)[i] {
          i2++
        } else {
          break
        }
      }
      if i2 == 10 {
        return false
      }
      cur_val += string((*x)[i])
      i++
    }
    if len(cur_val) > 3 || len(cur_val) == 0 {
      return false
    }
    i++
  }
  cur_val = ""
  for i < n {
    i2 = 0
    for i2 < 10 {
      if ref_nb[i2] != (*x)[i] {
        i2++
      } else {
        break
      }
    }
    if i2 == 10 {
      return false
    }
    cur_val += string((*x)[i])
    i++
  }
  if len(cur_val) > 3 || len(cur_val) == 0 {
    return false
  }
  return true
}

func GoodRoom(given_username string) bool {
  var i2 int
  var cur_val uint8
  for i := 0; i < len(given_username); i++ {
    cur_val = given_username[i]
    for i2 = 0; i2 < 2; i2++ {
      if cur_val == banned_char[i2] {
        return false
      }
    }
  }
  return true 
}

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
    for i2 < 23 && cur_val != ref_spechr[i2] {
      i2++
    }
    if i2 < 23 {
      agn = false
    }
    i++
  }
  if agn {
    return false
  }
  return true
}

func DisplayMsg(x *tview.TextView, 
                app *tview.Application, 
                conn *websocket.Conn,
                aes_key *string) {
  var msg []byte
  var err error
  var deciphered_msg string
  var str_msg string
  var str_msg2 string
  var i int
  for {
    _, msg, err = (*conn).ReadMessage()
    if err != nil {
      (*app).Stop()
      return
    }
    str_msg = string(msg)
    str_msg2 = ""

    i = 0
    for str_msg[i] != ' ' {
      i++
    }
    i++
    str_msg2 = str_msg[i:]
    
    deciphered_msg = decipherer(&str_msg2, aes_key)
    app.QueueUpdateDraw( func () {
                (*x).Write([]byte(str_msg[:i] + deciphered_msg))
                (*x).ScrollToEnd()
    })
  }
  return
}

func Quitting(conn *websocket.Conn, cur_aes_key *string) {
  var text string = "Disconnected \n"
  ciphered_text := cipherer(&text, cur_aes_key)
  conn.WriteMessage(websocket.TextMessage, []byte(ciphered_text))
  time.Sleep(200 * time.Millisecond)
  defer conn.Close()
}

func Int32ToString(x *int32) string {
  if *x == 0 {
    return "0"
  }
  const base int32 = 10
  var remainder int32
  rtn_str := ""
  for *x > 0 {
    remainder = *x % base
    rtn_str = string(remainder + 48) + rtn_str
    *x -= remainder
    *x /= 10
  }
  return rtn_str
}

func UTF8ToURL(x string) string {
  rtn_str := ""
  var vl int32
  for i := 0; i < len(x); i++ {
    vl = int32(x[i])
    rtn_str += Int32ToString(&vl)
    rtn_str += "-"
  }
  rtn_str = rtn_str[:len(rtn_str) - 1]
  return rtn_str
}

func main() {

  args := os.Args
  if len(args) < 2 {
    fmt.Println("server Ip and Ids missing")
    return
  }
  pre_formated := os.Args[1]
  var i int = 0
  ip_val := ""
  port_val := ""
  id_val := ""
  room_val := ""
  var n int = len(pre_formated)
  for i < n && pre_formated[i] != '@' {
    id_val += string(pre_formated[i])
    i++
  }
  if i == n || len(id_val) == 0 {
    fmt.Println("not valid")
    return
  }
  is_valid := GoodId(id_val)
  if !is_valid {
    fmt.Println("id not valid")
    return
  }
  i++
  for i < n && pre_formated[i] != ':' {
    ip_val += string(pre_formated[i])
    i++
  }
  if i == n || len(ip_val) == 0 {
    fmt.Println("not valid")
    return
  }
  is_valid = GoodIP(&ip_val)
  if !is_valid {
    fmt.Println("ip not valid")
    return
  }
  i++
  for i < n && pre_formated[i] != '/' {
    port_val += string(pre_formated[i])
    i++
  }
  if i == n || len(port_val) == 0 {
    fmt.Println("not valid")
    return
  }
  is_valid = GoodPort(port_val)
  if !is_valid {
    fmt.Println("port not valid")
    return
  }
  i++
  for i < n {
    room_val += string(pre_formated[i])
    i++
  }
  if len(room_val) == 0 {
    fmt.Println("not valid")
    return
  }
  is_valid = GoodRoom(room_val)
  if !is_valid {
    fmt.Println("room not valid")
    return
  }

  id_val_ciphered1 := ByteCipherer(&id_val, &global_aes_key)
  fmt.Println("0", id_val_ciphered1, len(id_val_ciphered1))

  id_val_ciphered := ""
  var tmp_int32 int32
  for i := 0; i < len(id_val_ciphered1); i++ {
    tmp_int32 = int32(id_val_ciphered1[i])
    id_val_ciphered += Int32ToString(&tmp_int32)
    id_val_ciphered += "-"
  }
  id_val_ciphered = id_val_ciphered[:len(id_val_ciphered) - 1]

  //rec2 := ByteDecipherer(&id_val_ciphered1, &global_aes_key)
  //fmt.Println("deciphered:", rec2)

  my_addr := ip_val + ":" + port_val + "/" + room_val + "_" + id_val_ciphered
  my_addr2 := ip_val + ":" + port_val + "/"

  data, err := os.ReadFile("pubKey.pem")
  if err != nil {
    fmt.Println(err)
    return
  }
  data_key := string(data)
  if data_key == "" {
    fmt.Println("Generating RSA keys...")
    private_key, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
      fmt.Println(err)
      return
    }
    public_key := &private_key.PublicKey
    private_key_bytes := x509.MarshalPKCS1PrivateKey(private_key)
    privatekeyPEM := pem.EncodeToMemory(
                     &pem.Block{Type: "RSA PRIVATE KEY", 
                     Bytes: private_key_bytes})
    err = os.WriteFile("privateKey.pem", 
                      privatekeyPEM, 0644)
    if err != nil {
      fmt.Print(err)
      return
    }
    public_key_bytes := x509.MarshalPKCS1PublicKey(public_key)
    if err != nil {
      fmt.Println(err)
      return
    }
    publickeyPEM := pem.EncodeToMemory(
                     &pem.Block{Type: "PUBLIC KEY", 
                     Bytes: public_key_bytes})
    err = os.WriteFile("pubKey.pem", 
                      publickeyPEM, 0644)
    if err != nil {
      fmt.Print(err)
      return
    }
  }

  data_placeholder := &bytes.Buffer{}
  writer := multipart.NewWriter(data_placeholder)
  part, err := writer.CreateFormFile("PubKey", "PubKey")
  if err != nil {
    fmt.Println(err)
    return
  }
  file, err := os.Open("pubKey.pem")
  if err != nil {
    fmt.Println(err)
    return
  }
  _, err = io.Copy(part, file)
  if err != nil {
    fmt.Println(err)
    return
  }
  err = writer.Close()
  if err != nil {
    fmt.Println(err)
    return
  }
  file.Close()
  clt := &http.Client{}
  req, err := http.NewRequest("POST", "http://" + my_addr, data_placeholder)
  if err != nil {
    fmt.Println(err)
    return
  }
  req.Header.Add("Content-Type", writer.FormDataContentType())
  resp, err := clt.Do(req)
  body, err := io.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err)
    return
  }
  resp.Body.Close()
  str_body := string(body)
  i = 0
  if str_body != "Ok" {
    for i < 8 {
      if str_body == ref_err_resp[i] {
        fmt.Println(str_body)
        return
      }
      i++
    }
    f, err := os.Create("my_rooms/" + room_val)
    if err != nil {
      fmt.Println(err)
      return
    }
    f.Close()
    data_private_key, err := os.ReadFile("privateKey.pem")
    if err != nil {
      fmt.Println(err)
      return
    }
    block, _ := pem.Decode(data_private_key)
    final_private_key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
      fmt.Println(err)
      return
    }
    decrypted_aes, err := rsa.DecryptPKCS1v15(rand.Reader, final_private_key, body)
    if err != nil {
      fmt.Println(err)
      return
    }
    err = os.WriteFile("my_rooms/" + room_val, 
         decrypted_aes, 
         0644)
    if err != nil {
      fmt.Println(err)
      return
    }
  }
  
  raw_aes_key, err := os.ReadFile("my_rooms/" + room_val)
  if err != nil {
    fmt.Println(err)
    return
  }
  cur_aes_key := string(raw_aes_key)

  conn, _, err := websocket.DefaultDialer.Dial("ws://" + my_addr2 + "chat/" + room_val + "_" + id_val_ciphered, nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  var text string = "Connected \n"
  ciphered_text := cipherer(&text, &cur_aes_key)
  conn.WriteMessage(websocket.TextMessage, []byte(ciphered_text))
  defer Quitting(conn, &cur_aes_key)
  //defer conn.Close()
 
  app := tview.NewApplication()

  custom_color := tcell.NewRGBColor(106, 
                                    44,
                                    43)

  custom_color2 := tcell.NewRGBColor(183, 
                                    81,
                                    79)

  custom_color3 := tcell.NewRGBColor(109, 
                                    133,
                                    201)

  info_box := tview.NewTextView().
    SetText("Escape to QUIT").
    SetTextColor(tcell.ColorYellow)

  info_box2 := tview.NewTextView().
    SetText("IP: | ChatRoom: ").
    SetTextColor(custom_color3)

  messageBox := tview.NewTextView().
    SetTextAlign(tview.AlignLeft).
    SetScrollable(true)

  messageBox.SetBorder(true).
    SetBorderColor(custom_color2)
    
  messageBox.SetTextColor(custom_color3)

  inputField := tview.NewInputField().
    SetLabel("Message: ").
    SetFieldWidth(90)
 
  inputField.SetLabelColor(tcell.ColorRed)
  inputField.SetFieldBackgroundColor(custom_color)
  inputField.SetFieldTextColor(tcell.ColorWhite)
  
  inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyEscape {
          
          app.Stop()
          return nil
        }
        if event.Key() == tcell.KeyEnter {
          text := inputField.GetText() + "\n"
          ciphered_text := cipherer(&text, &cur_aes_key)
          conn.WriteMessage(websocket.TextMessage, []byte(ciphered_text))
          inputField.SetText("") 
        }
        
        return event
  })

  flex := tview.NewFlex().
    SetDirection(tview.FlexRow).
    AddItem(info_box, 1, 0, false).
    AddItem(messageBox, 0, 3, false).
    AddItem(info_box2, 1, 0, false).
    AddItem(tview.NewBox(), 1, 0, false).
    AddItem(inputField, 2, 1, true)

 
  go DisplayMsg(messageBox, app, conn, 
                &cur_aes_key)

  err = app.SetRoot(flex, true).Run()
  if err != nil {
    fmt.Println(err)
    return
  }

}


