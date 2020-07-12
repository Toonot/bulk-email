package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/urfave/cli"
)

func main() {

	var templateFileName, recipientListFileName string

	//For reading the template and recipientList filepaths, cli is utilized
	// https://github.com/urfave/cli
	app := &cli.App{
		Name:  "bmail",
		Usage: "Send Bulk Emails",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "template, t",
				Usage:    "Load the template file (HTML)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "maillist, m",
				Usage:    "Load the maillist file (csv)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {

			templateFileName = c.String("template")
			recipientListFileName = c.String("maillist")
			fmt.Println("For help run bmail --help")
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	subject := "Test Mail"
	ReadRecipient(recipientListFileName, templateFileName, subject)

}

//Message struct
type Message struct {
	to      string
	from    string
	subject string
	body    string
}

//ParseTemplate parses the template
func ParseTemplate(templateFileName string, data interface{}) string {

	// Open the file
	tmpl, err := template.ParseFiles(templateFileName)
	if err != nil {
		log.Fatalln("Couldn't open the template", err)
	}
	buf := new(bytes.Buffer)
	tmpl.Execute(buf, data)
	return buf.String()
}

//ReadRecipient reads list of recipients from csv file
func ReadRecipient(recipientListFileName, templateFileName, subject string) {

	// Open the file
	csvFile, err := os.Open(recipientListFileName)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	// Parse the file
	reader := csv.NewReader(bufio.NewReader(csvFile))

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		//validating email structure using regex
		erre := ValidateFormat(record[1])
		if erre != nil {
			fmt.Println(record[1], "Email is not valid.")
		}
		if erre == nil {
			//Structure for sending data to
			data := struct {
				Name string
			}{
				Name: record[0],
			}
			//Parsing data to template (i.e "Name" in place of {.Name})
			body := ParseTemplate(templateFileName, data)
			m := Message{
				to:      record[1],
				subject: subject,
				body:    body,
				from:    "mail@example.com",
			}
			m.Send()
			time.Sleep(500 * time.Millisecond)

		}
	}
}

//Send for sending email
func (m *Message) Send() {
	// Set up authentication information.
	auth := smtp.PlainAuth("", "sender@example.org", "password", "localhost")

	//Convert "to" to []string
	to := []string{m.to}
	//RFC 822-style email format
	//Omit "to" parameter in msg to send as bcc
	msg := []byte("From: " + m.from + "\r\n" +
		"To: " + m.to + "\r\n" +
		"Subject: " + m.subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" +
		m.body + "\r\n")

	err := smtp.SendMail("localhost:1025", auth, "sender@example.org", to, msg)
	count := 0
	for err != nil && count <= 10 {
		err = smtp.SendMail("localhost:1025", auth, "sender@example.org", to, msg)
		count++
	}
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed sending to ", m.to)
	} else {
		fmt.Println("Email Sent to ", m.to)
	}

}

// ValidateFormat validates the email using regex
func ValidateFormat(email string) error {
	regex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !regex.MatchString(email) {
		return errors.New("Invalid Format")
	}
	return nil
}
