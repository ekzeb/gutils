package util

import (
	"gopkg.in/gomail.v2"
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"html/template"
	"bytes"
)

var parsedEmailTemplates map[string]*template.Template

func SendHtmlEmail(serverAddress string, pass string, from string, subject string, templates []string, params map[string]interface{} , to ...string) (err error)  {

	host, portStr, er := net.SplitHostPort(serverAddress)

	if er != nil {
		err = er
		log.Printf("Error in Email server config %v err: %v", serverAddress, err)
		return
	}

	if port, er := strconv.Atoi(portStr); er == nil {
                if parsedEmailTemplates == nil {
			parsedEmailTemplates = make(map[string]*template.Template)
		}
		if parsedEmailTemplates[templates[0]] == nil {
			parsedEmailTemplates[templates[0]] = ParseTemplate(templates, nil)
		}
		var body bytes.Buffer
		parsedEmailTemplates[templates[0]].Execute(&body, params)
		m := gomail.NewMessage()
		m.SetHeader("From", from)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body.String())
		//m.Attach("/home/ekzeb/image.jpg")

		d := gomail.NewDialer(host, port, from, pass)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		if err = d.DialAndSend(m); err != nil {
			log.Printf("Error sending Email %v", body)
		}
	} else {
		err = er
		log.Printf("Error in Email server config no port '%v' err: %v", portStr, err)
		return
	}

	return
}


func SendEmail(serverAddress string, pass string, from string, subject, body string, to ...string) (err error)  {

	host, portStr, er := net.SplitHostPort(serverAddress)

	if er != nil {
		err = er
		log.Printf("Error in Email server config %v err: %v", serverAddress, err)
		return
	}

	if port, er := strconv.Atoi(portStr); er == nil {

		m := gomail.NewMessage()
		m.SetHeader("From", from)
		m.SetHeader("To", to...)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)

		d := gomail.NewDialer(host, port, from, pass)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		if err = d.DialAndSend(m); err != nil {
			log.Printf("Error sending Email %v", body)
		}
	} else {
		err = er
		log.Printf("Error in Email server config no port '%v' err: %v", portStr, err)
		return
	}

	return
}