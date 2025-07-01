package main

// import (
// 	"encoding/json"
// )

// func InboxHandler(body []byte) []byte {
// 	inbox, err := GetInbox()
// 	if err != nil {
// 		return http.ErrorJSON(err.Error())
// 	}
// 	return http.JSON(inbox)
// }

// func SendHandler(body []byte) []byte {
// 	var email types.Email
// 	if err := json.Unmarshal(body, &email); err != nil {
// 		return http.ErrorJSON("Invalid input")
// 	}
// 	err := smtpclient.Send(email)
// 	if err != nil {
// 		return http.ErrorJSON(err.Error())
// 	}
// 	return http.JSON(map[string]string{"status": "sent"})
// }
