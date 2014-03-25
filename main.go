package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

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
}

func main() {
	data, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Fatalln(err)
	}

	var c Config

	err = json.Unmarshal(data, &c)

	conn, err := gobeanstalk.Dial(c.Address)

	if err != nil {
		log.Fatalln(err)
	}

	mg_client := mailgun.NewClient(c.Key, c.Mailbox)

	for {
		j, err := conn.Reserve()

		if err != nil {
			log.Println("reserve failed")
			log.Fatal(err)
		}

		err = conn.Delete(j.Id)

		if err != nil {
			log.Fatal(err)
		}

		message := mailgun.Message{
			FromName:    c.FromName,
			FromAddress: c.FromAddress,
			ToAddress:   string(j.Body),
			Subject:     c.Subject,
			Body:        c.Body,
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
