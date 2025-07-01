package main

// import (
// 	"fmt"

// 	smtp "github.com/loredanacirstea/wasmx-env-smtp"
// )

// func Send(email EmailRecord) error {
// 	req := &smtp.SmtpSendMailRequest{
// 		From:    email.From,
// 		To:      []string{email.To},
// 		Subject: email.Subject,
// 		Body:    email.Body,
// 	}
// 	resp := smtp.Send(req)
// 	if resp.Error != "" {
// 		return fmt.Errorf(resp.Error)
// 	}
// 	return nil
// }
