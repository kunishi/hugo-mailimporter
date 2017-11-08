package main

import (
  "bufio"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "mime"
  "net/mail"
  "os"
  "regexp"
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

func main() {
  type FrontMatter struct {
    From string `toml:"from"`
    Title string `toml:"title"`
    Date string `toml:"date"`
  }

  // https://stackoverflow.com/questions/35097318/email-subject-header-decoding-in-different-charset-like-iso-2022-jp-gb-2312-e
  CharsetReader := func (label string, input io.Reader) (io.Reader, error) {
    label = strings.Replace(label, "windows-", "cp", -1)
    encoding, _ := charset.Lookup(label)
    return encoding.NewDecoder().Reader(input), nil
  }

  reader := bufio.NewReader(os.Stdin)
  msg, _ := email.ParseMessage(reader)
  dec := mime.WordDecoder{CharsetReader: CharsetReader}

  fmt.Println("---")
  from, _ := dec.DecodeHeader(msg.Header["From"][0])
  title, _ := dec.DecodeHeader(msg.Header["Subject"][0])
  date, _ := mail.ParseDate(msg.Header["Date"][0])
  date_str := date.Format(time.RFC3339)
  fm := FrontMatter{From: from, Title: title, Date: date_str}
  b, err := toml.Marshal(fm)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(string(b) + "---")

  content_type := msg.Header["Content-Type"][0]
  r := regexp.MustCompile(`[Ii][Ss][Oo]-2022-[Jj][Pp]`)
  if r.MatchString(content_type) {
    body, _ := jis_to_utf8(string(msg.Body))
    fmt.Println(body)
  } else {
    fmt.Println(string(msg.Body))
  } 
}
