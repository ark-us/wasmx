package vmimap

import (
	"fmt"
	"io"
	"strings"

	"cosmossdk.io/log"
	"github.com/emersion/go-message/mail"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// connectToIMAP establishes a connection to the IMAP server
func connectToIMAP(imapServer string, username, password string, options *imapclient.Options) (*imapclient.Client, error) {
	// Connect to IMAP server over TLS.
	// username-password credentials or OAuth2 or an app password
	c, err := imapclient.DialTLS(imapServer, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}

	if err := c.Login(username, password).Wait(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to login: %v", err)
	}

	return c, nil
}

// connectToIMAP establishes a connection to the IMAP server
func connectToIMAPOauth2(imapServerUrl string, username string, accessToken string, options *imapclient.Options) (*imapclient.Client, error) {
	// Connect to the IMAP server
	c, err := imapclient.DialTLS(imapServerUrl, options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	xauth := &OAuth2Authenticator{username: username, accessToken: accessToken}

	// Authenticate using OAuth2
	// auth := imapoauth2.NewOAuthBearerAuthenticator(accessToken)
	if err := c.Authenticate(xauth); err != nil {
		return nil, fmt.Errorf("failed to authenticate with oauth2: %v", err)
	}
	return c, nil
}

func fetchEmailIds(c *imapclient.Client, folder *imap.SelectData, username string, filters FetchFilter) (imap.NumSet, uint32, error) {
	var uidSet imap.NumSet
	var count uint32
	var criteria *imap.SearchCriteria = nil
	var err error
	limit := filters.Limit
	start := filters.Start
	numMsg := folder.NumMessages
	var uids *imap.SearchData

	criteria, err = buildSearchCriteria(filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build search criteria: %v", err.Error())
	}

	if criteria != nil {
		uids, err = c.UIDSearch(criteria, nil).Wait()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to search emails: %v", err)
		}

		// Fetch only the matching UIDs
		uidSet = uids.All
		count = uint32(len(uids.AllUIDs()))
		return uidSet, count, nil
	}

	// empty criteria should return all emails, but the library throws an error
	var uidNums []imap.UID
	max := start + limit
	count = limit
	if start > numMsg {
		return nil, 0, fmt.Errorf("imap: invalid start criteria")
	}
	if max > numMsg {
		max = numMsg
		count = numMsg - start + 1
	}
	for i := start; i <= max; i++ {
		uidNums = append(uidNums, imap.UID(i))
	}
	uidSet = imap.UIDSetNum(uidNums...)
	return uidSet, count, nil
}

func imapFetch(c *imapclient.Client, logger log.Logger, numSet imap.NumSet, options *imap.FetchOptions, bodySection *imap.FetchItemBodySection) ([]Email, error) {
	var emails []Email

	// Fetch command
	fetchCmd := c.Fetch(numSet, options)

	for {
		msgd := fetchCmd.Next()
		if msgd == nil {
			break
		}

		// TODO for big attachment support use Next()
		msg, err := msgd.Collect()
		if err != nil {
			logger.Info("Failed to collect email for UID %d: %v", msg.UID, err)
			continue
		}

		// fmt.Printf("UID: %d, Subject: %s, From: %v\n", msg.UID, msg.Envelope.Subject, msg.Envelope.From)

		// just find the header
		// header := msg.FindBodySection(bodySection)
		// logger.Info("Header:\n%v", string(header))

		// data := msg.FindBodySection(bodySection)
		// emails = append(emails, string(data))

		// Get raw email content
		raw := msg.FindBodySection(bodySection)

		// Parse headers and body using go-message
		mr, err := mail.CreateReader(strings.NewReader(string(raw)))
		if err != nil {
			logger.Info("Failed to parse email for UID %d: %v", msg.UID, err)
			continue
		}

		header := make(map[string][]string)
		fields := mr.Header.Fields()
		bh := "" // body hash
		for {
			more := fields.Next()
			if !more {
				break
			}

			key := fields.Key()
			value := fields.Value()
			header[key] = append(header[key], value)
			if key == "Dkim-Signature" {
				parts := strings.Split(value, "; ")
				for _, p := range parts {
					if p[0:3] == "bh=" {
						bh = p[3:]
					}
				}
			}
		}

		body, attachments := extractEmailParts(logger, mr, msg)

		// Populate Email struct
		email := Email{
			Raw:          string(raw),
			UID:          msg.UID,
			Flags:        msg.Flags,
			InternalDate: msg.InternalDate,
			RFC822Size:   msg.RFC822Size,
			Envelope:     msg.Envelope,
			// BodyStructure: &msg.BodyStructure,
			Header:      header,
			Body:        body,
			Attachments: attachments,
			Bh:          bh,
		}

		emails = append(emails, email)
	}

	// After iteration, check for any errors:
	if err := fetchCmd.Close(); err != nil {
		fmt.Printf("Error during fetch: %v\n", err)
	}

	return emails, nil
}

// extractEmailParts extracts the body (first text part) and attachments from a mail.Reader.
func extractEmailParts(logger log.Logger, mr *mail.Reader, msg *imapclient.FetchMessageBuffer) (body string, attachments []Attachment) {
	attachments = []Attachment{}

	// Iterate through each part of the message.
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Info("Failed to read part for UID %d: %v", msg.UID, err)
			break
		}

		// If the part is inline (text content)
		if inlineHeader, ok := part.Header.(*mail.InlineHeader); ok {
			contentType, _, _ := inlineHeader.ContentType()
			if contentType == "text/plain" || contentType == "text/html" {
				bodyBytes, err := io.ReadAll(part.Body)
				if err != nil {
					logger.Info("Failed to read body for UID %d: %v", msg.UID, err)
					continue
				}
				// Save the first text part as body (or you can concatenate if you expect multiple)
				if body == "" {
					body = string(bodyBytes)
				}
			}
		}

		// If the part is an attachment
		if attachHeader, ok := part.Header.(*mail.AttachmentHeader); ok {
			// Try to get the filename; if not available, set a default.
			filename, err := attachHeader.Filename()
			if err != nil {
				filename = "unknown"
			}
			contentType, _, _ := attachHeader.ContentType()
			data, err := io.ReadAll(part.Body)
			if err != nil {
				logger.Info("Failed to read attachment for UID %d: %v", msg.UID, err)
				continue
			}
			attachments = append(attachments, Attachment{
				Filename:    filename,
				ContentType: contentType,
				Data:        data,
			})
		}
	}
	return
}

func buildSearchCriteria(filters FetchFilter) (*imap.SearchCriteria, error) {
	criteria := filters.Search

	// Header-based filters (From, To, Subject)
	if filters.From != "" {
		if criteria == nil {
			criteria = &imap.SearchCriteria{}
		}
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "From",
			Value: filters.From,
		})
	}
	if filters.To != "" {
		if criteria == nil {
			criteria = &imap.SearchCriteria{}
		}
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "To",
			Value: filters.To,
		})
	}
	if filters.Subject != "" {
		if criteria == nil {
			criteria = &imap.SearchCriteria{}
		}
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{
			Key:   "Subject",
			Value: filters.Subject,
		})
	}

	// Text-based filter
	if filters.Content != "" {
		if criteria == nil {
			criteria = &imap.SearchCriteria{}
		}
		criteria.Text = []string{filters.Content}
	}
	return criteria, nil
}
