package main

import (
  "bufio"
  "crypto/md5"
  "encoding/hex"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "mime"
  "net/mail"
  "os"
  "strings"
  "time"
  "github.com/pelletier/go-toml"
  "github.com/veqryn/go-email/email"
  "golang.org/x/net/html/charset"
  "golang.org/x/text/encoding/japanese"
  "golang.org/x/text/transform"
)

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

func main() {
  type FrontMatter struct {
    From string `toml:"from"`
    Title string `toml:"title"`
    Date string `toml:"date"`
    PostId string `toml:"post_id"`
  }

  // https://stackoverflow.com/questions/35097318/email-subject-header-decoding-in-different-charset-like-iso-2022-jp-gb-2312-e
  CharsetReader := func (label string, input io.Reader) (io.Reader, error) {
    label = strings.Replace(label, "windows-", "cp", -1)
    encoding, _ := charset.Lookup(label)
    return encoding.NewDecoder().Reader(input), nil
  }

  reader := bufio.NewReader(os.Stdin)
  msg, _ := email.ParseMessage(reader)
  messageBody := string(msg.Body)
  for _, part := range msg.MessagesAll() {
    mediaType, params, err := part.Header.ContentType()
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println(mediaType, params)
    switch mediaType {
    case "text/plain":
      messageBody = string(part.Body)
      charset, ok := params["charset"]
      if ok {
        switch charset {
        case "iso-2022-jp":
          messageBody, _ = jis_to_utf8(messageBody)
        }
      }
    }
  }

  post_id := ""
  _, ok := msg.Header["Message-Id"]
  if ok {
    message_id := msg.Header["Message-Id"][0]
    post_id = GetMD5Hash(message_id)
  } else {
    post_id = GetMD5Hash(string(msg.Body))
  }

  dec := mime.WordDecoder{CharsetReader: CharsetReader}
  fmt.Println("---")
  from, _ := dec.DecodeHeader(msg.Header["From"][0])
  title, _ := dec.DecodeHeader(msg.Header["Subject"][0])
  date, _ := mail.ParseDate(msg.Header["Date"][0])
  date_str := date.Format(time.RFC3339)
  fm := FrontMatter{From: from, Title: title, Date: date_str, PostId: post_id}
  b, err := toml.Marshal(fm)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(string(b) + "---")
  fmt.Println("<pre>")
  fmt.Println(messageBody)
  fmt.Println("</pre>")
}
