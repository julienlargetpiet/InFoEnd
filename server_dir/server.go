package main

import (
  "fmt"
  "github.com/gdamore/tcell/v2"
  "github.com/rivo/tview"
  "database/sql"
  "time"
  _"github.com/go-sql-driver/mysql"
)

var idx_input int = 0
var ref_ltr = [52]uint8{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var ref_nb = [10]uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var ref_spechr = [24]uint8{'!', '.', ':', ';', '\\', '-', '%', '*', ',', '_', '/', '<', '>', '=', '[', ']', '\'', '{', '}', '[', ']', '(', ')', '"'}
var banned_char = [2]uint8{'/', ' '}

func ConnectDatabase() (*sql.DB, error) {
  var credentials = "kvv:1234@(localhost:3306)/InFoEnd"
  db, err := sql.Open("mysql", credentials)
  if err != nil {
    return nil, err
  }
  return db, nil
}

func GenerateAES256() string {
  rtn_str := ""
  var tm int64 = time.Now().Unix()
  for i := 0; i < 32; i++ {
    if tm % 3 == 0 {
      rtn_str += string(ref_ltr[tm % 52])
    } else if tm % 2 == 0 {
      rtn_str += string(ref_nb[tm % 10])
    } else {
      rtn_str += string(ref_spechr[tm % 24])
    }
    tm /= 2
  }
  return rtn_str
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

func CreateRoom(srv_room string, usernames string, ids string) (bool, string, []string, []string) {
  is_valid := GoodRoom(srv_room)
  var rtn_usernames []string
  var rtn_ids []string
  if !is_valid {
    return false, "Bad Room Name", rtn_usernames, rtn_usernames
  }
  var i int = 0
  var i2 int
  var ln_usrs int
  var ln_ids int
  cur_val := ""
  usernames += ","
  ids += ","
  var n int = len(usernames)
  for i < n {
    if usernames[i] != ',' {
      cur_val += string(usernames[i])
    } else {
      if cur_val == "" {
        return false, "An username is empty", rtn_usernames, rtn_usernames
      }
      is_valid = GoodRoom(cur_val)
      if !is_valid {
        return false, "An username is not valid", rtn_usernames, rtn_usernames
      }
      rtn_usernames = append(rtn_usernames, cur_val)
      cur_val = ""
    }
    i++
  }
  ln_usrs = len(rtn_usernames)
  for i = 0; i < ln_usrs; i++ {
    cur_val = rtn_usernames[i]
    for i2 = 0; i2 < ln_usrs; i2++ {
      if i != i2 {
        if cur_val == rtn_usernames[i2] {
          return false, "No same usernames allowed", rtn_usernames, rtn_usernames
        }
      }
    }
  }
  cur_val = ""
  var n2 int = len(ids)
  i = 0
  for i < n2 {
    if ids[i] != ',' {
      cur_val += string(ids[i])
    } else {
      if len(cur_val) < 5 {
        return false, "An id is not of length superior to 4", rtn_usernames, rtn_usernames
      }
      is_valid = GoodRoom(cur_val)
      if !is_valid {
        return false, "An id is not valid", rtn_usernames, rtn_usernames
      }
      rtn_ids = append(rtn_ids, cur_val)
      cur_val = ""
    }
    i++
  }
  ln_ids = len(rtn_ids)
  for i = 0; i < ln_ids; i++ {
    cur_val = rtn_ids[i]
    for i2 = 0; i2 < ln_ids; i2++ {
      if i != i2 {
        if cur_val == rtn_ids[i2] {
          return false, "No same ids allowed", rtn_usernames, rtn_usernames
        }
      }
    }
  }
  if ln_usrs > ln_ids {
    return false, "Too much usernames compared to ids", rtn_usernames, rtn_usernames
  } else if ln_usrs < ln_ids {
    return false, "Not enough much usernames compared to ids", rtn_usernames, rtn_usernames
  }
  return true, "", rtn_usernames, rtn_ids
}

func main() {

  db , err := ConnectDatabase()
  if err != nil {
    fmt.Println(err)
    return
  }

  var is_valid bool
  var resp string
  var srv_room string
  var usernames string
  var ids string
  var rtn_ids []string
  var rtn_usernames []string
  var i int

  app := tview.NewApplication()

  custom_color := tcell.NewRGBColor(106, 
                                    44,
                                    43)

  top_label := tview.NewTextView().
    SetLabel("    * Usernames and ids are separated by ',' Use TAB to navigate")

  top_label.SetTextColor(tcell.ColorYellow)

  top_label2 := tview.NewTextView().
    SetText("    * Down Arrow Key to be on Submit Button")
   
  top_label2.SetTextColor(tcell.ColorYellow)

  top_label3 := tview.NewTextView().
    SetText("    * Enter to CREATE and Escape to QUIT")
 
  top_label3.SetTextColor(tcell.ColorYellow)

  err_label := tview.NewTextView().
    SetText("")

  err_label.SetTextColor(tcell.ColorOrange)

  input1 := tview.NewInputField().
    SetLabel("  Server_Room: ").
    SetFieldWidth(30).
    SetLabelColor(tcell.ColorRed)

  input1.SetFieldBackgroundColor(custom_color)

  input2 := tview.NewInputField().
    SetLabel("    Usernames: ").
    SetFieldWidth(30).
    SetLabelColor(tcell.ColorRed)

  input2.SetFieldBackgroundColor(custom_color)

  input3 := tview.NewInputField().
    SetLabel("          Ids: ").
    SetFieldWidth(30).
    SetLabelColor(tcell.ColorRed)

  input3.SetFieldBackgroundColor(custom_color)

  inputs := [3]*tview.InputField{input1, input2, input3}

  app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
     app.Stop()
      return nil
    }
    if event.Key() == tcell.KeyTab {
      idx_input += 1
      app.SetFocus(inputs[idx_input % 3])
    }
    if event.Key() == tcell.KeyEnter {
      srv_room = input1.GetText()
      usernames = input2.GetText()
      ids = input3.GetText()
      is_valid, resp, rtn_usernames, rtn_ids = CreateRoom(srv_room, 
                                                          usernames, 
                                                          ids)
      if !is_valid {
        err_label.SetText(resp)
        return event
      }
      aes_key := GenerateAES256()
      if err != nil {
        err_label.SetText("Error generating AES key")
        return event
      }
      _, err = db.Exec("INSERT INTO Status VALUES(?, 0, ?);", 
                                                   srv_room,
                                                   aes_key)
      if err != nil {
        str_err := err.Error()
        err_label.SetText(str_err)
        return event
      }
      _, err = db.Exec("CREATE TABLE " + srv_room + " (username VARCHAR(15), id VARCHAR(16), already BOOL);")
      if err != nil {
        str_err := err.Error()
        err_label.SetText(str_err)
        return event
      }
      i = 0
      for i < len(rtn_usernames) {
        _, err = db.Exec("INSERT INTO " + srv_room + " VALUES(?, ?, 0);", 
                          rtn_usernames[i], rtn_ids[i])
        if err != nil {
          break
        }
        i++
      }
      if i == len(rtn_usernames) {
        err_label.SetText("Server Room created successfully")
      } else {
        err_label.SetText("Something went wrong")
      }
      return event
    }

    return event
  })

  flex := tview.NewFlex().
    SetDirection(tview.FlexRow).
    AddItem(tview.NewBox(), 1, 2, false).
    AddItem(top_label, 1, 3, false).
    AddItem(top_label2, 1, 3, false).
    AddItem(top_label3, 1, 3, false).
    AddItem(tview.NewBox(), 1, 1, false).
    AddItem(input1, 1, 3, true).
    AddItem(tview.NewBox(), 1, 1, false).
    AddItem(input2, 1, 3, false).
    AddItem(tview.NewBox(), 1, 1, false).
    AddItem(input3, 1, 3, true).
    AddItem(tview.NewBox(), 1, 1, false).
    AddItem(tview.NewBox(), 1, 1, false).
    AddItem(err_label, 1, 1, false).
    AddItem(tview.NewBox(), 1, 1, false)


  err = app.SetRoot(flex, true).Run()
  if err != nil {
    fmt.Println(err)
    return
  }

}


