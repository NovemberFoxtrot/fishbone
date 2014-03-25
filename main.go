package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iwanbk/gobeanstalk"
	"github.com/sirsean/go-mailgun/mailgun"
)

type Config struct {
	Address     string `json:"address"`
	Body        string `json:"body"`
	FromAddress string `json:"fromaddress"`
	FromName    string `json:"fromname"`
	Key         string `json:"key"`
	Mailbox     string `json:"mailbox"`
	Subject     string `json:"subject"`
	Database    string `json:"database"`
	Password    string `json:"password"`
	Username    string `json:"username"`
}

func main() {
	data, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Fatalln(err)
	}

	var c Config

	err = json.Unmarshal(data, &c)

	db, err := sql.Open("mysql", c.Username+":"+c.Password+"@/"+c.Database)

	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	conn, err := gobeanstalk.Dial(c.Address)

	if err != nil {
		log.Fatalln(err)
	}

	mg_client := mailgun.NewClient(c.Key, c.Mailbox)

	for {
		j, err := conn.Reserve()

		if err != nil {
			log.Fatalln("reserve failed", err)
		}

		err = conn.Delete(j.Id)

		if err != nil {
			log.Fatalln(err)
		}

		email := string(j.Body)

		t := time.Now()
		t.Format("2006-01-02 15:04:05")

		result, err := db.Exec(`INSERT INTO email(email, created_at) VALUES(?,?);`, email, t)

		if err != nil {
			log.Println("db error", err, result)
		}

		message := mailgun.Message{
			Body:        c.Body,
			FromAddress: c.FromAddress,
			FromName:    c.FromName,
			Subject:     c.Subject,
			ToAddress:   email,
		}

		log.Println("Attempting to send to ", mg_client.Endpoint(message))

		body, err := mg_client.Send(message)

		if err != nil {
			log.Println("Got an error:", err)
		} else {
			log.Println(body)
		}
	}
}
