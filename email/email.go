package email

import (
	"fmt"
	"net/smtp"
)

func Send(smtpAddr, from, to, subject, body string) error {
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	))

	return smtp.SendMail(smtpAddr, nil, from, []string{to}, msg)
}
