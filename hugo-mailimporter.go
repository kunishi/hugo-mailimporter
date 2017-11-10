package main

import (
  "crypto/md5"
  "encoding/base64"
  "encoding/hex"
  "fmt"
  "io/ioutil"
  "log"
  "mime"
  "net/mail"
  "os"
  "path/filepath"
  "strings"
  "time"
  "github.com/pelletier/go-toml"
  "github.com/jhillyerd/enmime"
  "golang.org/x/text/encoding/japanese"
  "golang.org/x/text/transform"
)

type Attachment struct {
  Name string `toml:"name"`
  FileName string `toml:"filename"`
}

type FrontMatter struct {
  From string `toml:"from"`
  Title string `toml:"title"`
  Date string `toml:"date"`
  PostId string `toml:"post_id"`
  Attachments map[string]Attachment
}

func jis_to_utf8(str string) (string, error) {
  iostr := strings.NewReader(str)
  rio := transform.NewReader(iostr, japanese.ISO2022JP.NewDecoder())
  ret, err := ioutil.ReadAll(rio)
  if err != nil {
    return "", err
  }
  return string(ret), err
}

func GetMD5Hash(text string) string {
  hasher := md5.New()
  hasher.Write([]byte(text))
  return hex.EncodeToString(hasher.Sum(nil))
}

func MailConverter(aMail string) (string) {
  r, err := os.Open(aMail)
  if err != nil {
    log.Fatal(err)
  }

  msg, err := enmime.ReadEnvelope(r)
  if err != nil {
    log.Fatal(err)
  }

  attachments := make(map[string]Attachment)
  for _, attach := range msg.Attachments {
    content, _ := ioutil.ReadAll(attach)
    attach_id := GetMD5Hash(base64.StdEncoding.EncodeToString(content))

    filename := ""
    ext, _ := mime.ExtensionsByType(attach.ContentType)
    if ext != nil {
      if attach.ContentType == "application/octet-stream" && filepath.Ext(attach.FileName) == ".pdf" {
        filename = attach_id + ".pdf"
      } else {
        filename = attach_id + ext[0]
      }
      attachments[attach_id] = Attachment{Name: attach.FileName, FileName: filename}
    }
    //err := ioutil.WriteFile(attach.FileName, content, 0644)
    //if err != nil {
    //  log.Fatal(err)
    //}
  }

  post_id := ""
  message_id := msg.GetHeader("Message-Id")
  if message_id != "" {
    post_id = GetMD5Hash(message_id)
  } else {
    post_id = GetMD5Hash(msg.Text)
  }

  result := ""
  result += fmt.Sprintln("---")

  from := msg.GetHeader("From")
  title := msg.GetHeader("Subject")
  date, _ := mail.ParseDate(msg.GetHeader("Date"))
  date_str := date.Format(time.RFC3339)
  fm := FrontMatter{
    From: from,
    Title: title,
    Date: date_str,
    PostId: post_id,
    Attachments: attachments,
  }
  b, err := toml.Marshal(fm)
  if err != nil {
    log.Fatal(err)
  }
  result += fmt.Sprintf("%s", string(b))
  result += fmt.Sprintln("---")

  if len(msg.HTML) != 0 {
    result += fmt.Sprintln(msg.HTML)
  } else {
    result += fmt.Sprintln("<pre>")
    if msg.GetHeader("MIME-Version") != "" {
      result += fmt.Sprintln(msg.Text)
    } else {
      string_jis, _ := jis_to_utf8(msg.Text)
      result += fmt.Sprintln(string_jis)
    }
    result += fmt.Sprintln("</pre>")
  }

  return result
}

func main() {
  for _, mail := range os.Args[1:] {
    fmt.Print(MailConverter(mail))
  }
}
