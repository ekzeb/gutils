package util

import (
	"gopkg.in/gomail.v2"
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"html/template"
	"bytes"
	"regexp"
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
			parsedEmailTemplates[templates[0]] = ParseTemplate(templates...)
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

var emailRegexp = regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

func IsValidEmail(str string) bool {
	return emailRegexp.MatchString(str)
}