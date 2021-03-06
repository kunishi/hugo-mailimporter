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
  "github.com/PuerkitoBio/goquery"
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
  Type string `toml:"type"`
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

func HTMLBodyExtractor(html string) (string) {
  doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
  content, _ := doc.Find("body").Html()
  return content
}

func MailConverter(aMail string, config *toml.Tree) (string, string) {
  basedir := config.Get("basedir").(string)
  content_dir := basedir + config.Get("content_dir").(string)
  assets_dir := basedir + config.Get("assets_dir").(string)

  r, err := os.Open(aMail)
  if err != nil {
    log.Fatal(err)
  }

  msg, err := enmime.ReadEnvelope(r)
  if err != nil {
    log.Fatal(err)
  }

  post_id := ""
  message_id := msg.GetHeader("Message-Id")
  if message_id != "" {
    post_id = GetMD5Hash(message_id)
  } else {
    post_id = GetMD5Hash(msg.Text)
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

      attach_dir := assets_dir + "/" + post_id
      os.MkdirAll(attach_dir, os.ModePerm)
      err := ioutil.WriteFile(attach_dir + "/" + filename, content, 0644)
      if err != nil {
        log.Fatal(err)
      }
    }
  }

  result := ""
  result += fmt.Sprintln("+++")

  from := msg.GetHeader("From")
  title := msg.GetHeader("Subject")
  date, _ := mail.ParseDate(msg.GetHeader("Date"))
  date_str := date.Format(time.RFC3339)
  fm := FrontMatter{
    From: from,
    Title: title,
    Date: date_str,
    PostId: post_id,
    Type: config.Get("type").(string),
    Attachments: attachments,
  }
  b, err := toml.Marshal(fm)
  if err != nil {
    log.Fatal(err)
  }
  result += fmt.Sprintf("%s", string(b))
  result += fmt.Sprintln("+++")

  if len(msg.HTML) != 0 {
    result += fmt.Sprintln(HTMLBodyExtractor(msg.HTML))
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

  os.MkdirAll(content_dir, os.ModePerm)
  f, err := os.Create(content_dir + "/" + post_id + ".md")
  if err != nil {
    log.Fatal(err)
  }
  _, err = f.WriteString(result)
  if err != nil {
    log.Fatal(err)
  }
  
  return post_id, result
}

func main() {
  config, err := toml.LoadFile("config.toml")
  if err != nil {
    log.Fatal(err)
  }
  hugoConfig := config.Get("Hugo").(*toml.Tree)

  for _, mail := range os.Args[1:] {
    post_id, _ := MailConverter(mail, hugoConfig)
    fmt.Println(mail, post_id)
  }
}
