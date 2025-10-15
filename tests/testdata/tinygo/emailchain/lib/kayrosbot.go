package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	httpclient "github.com/loredanacirstea/wasmx-env-httpclient/lib"
	vmimap "github.com/loredanacirstea/wasmx-env-imap/lib"
)

const (
	BotEmail      = "bot@dmail.provable.dev"
	KayrosAPIURL  = "https://kayros.provable.dev/api/grpc/single-hash"
	DataTypeEmail = "70726f7661626c655f656d61696c000000000000000000000000000000000000"
)

// KayrosRequest represents the request to Kayros API
type KayrosRequest struct {
	DataItem string `json:"data_item"`
	DataType string `json:"data_type"`
}

// KayrosResponse represents the response from Kayros API
type KayrosResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// KayrosProof represents the JSON file structure attached to emails
type KayrosProof struct {
	Data   EmailDataProof `json:"data"`
	Kayros interface{}    `json:"kayros"`
}

// EmailDataProof represents the original email information
type EmailDataProof struct {
	From      []string `json:"from"`
	To        []string `json:"to"`
	Subject   string   `json:"subject,omitempty"`
	Timestamp int64    `json:"timestamp"`
	Hash      string   `json:"hash"`
}

// IsBotEmail checks if the email is addressed to the bot
func IsBotEmail(toAddresses []string) bool {
	for _, to := range toAddresses {
		if strings.EqualFold(strings.TrimSpace(to), BotEmail) {
			return true
		}
	}
	return false
}

// HashEmail computes SHA256 hash of email raw bytes and returns hex string
func HashEmail(emailRaw []byte) string {
	hash := sha256.Sum256(emailRaw)
	return hex.EncodeToString(hash[:])
}

// QueryKayros makes a POST request to Kayros API with email hash
func QueryKayros(emailHash string) (interface{}, error) {
	// Create request body
	reqBody := KayrosRequest{
		DataItem: emailHash,
		DataType: DataTypeEmail,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	httpReq := httpclient.HttpRequestWrap{
		Request: httpclient.HttpRequest{
			Method: "POST",
			Url:    KayrosAPIURL,
			Header: headers,
			Data:   reqBodyBytes,
		},
		ResponseHandler: httpclient.ResponseHandler{
			MaxSize: 10 * 1024 * 1024, // 10MB max response
		},
	}

	// Make the HTTP request
	resp := httpclient.Request(&httpReq)
	if resp.Error != "" {
		return nil, fmt.Errorf("HTTP request failed: %s", resp.Error)
	}

	// Check status code
	if resp.Data.StatusCode != 200 {
		return nil, fmt.Errorf("Kayros API returned status %d: %s", resp.Data.StatusCode, string(resp.Data.Data))
	}

	// Parse response as generic interface
	var kayrosResp interface{}
	err = json.Unmarshal(resp.Data.Data, &kayrosResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kayros response: %w", err)
	}

	return kayrosResp, nil
}

// CreateKayrosProof creates a proof structure combining email data and Kayros response
func CreateKayrosProof(from []string, to []string, subject string, emailRaw []byte, timestamp int64, kayrosResp interface{}) *KayrosProof {
	emailHash := HashEmail(emailRaw)

	return &KayrosProof{
		Data: EmailDataProof{
			From:      from,
			To:        to,
			Subject:   subject,
			Timestamp: timestamp,
			Hash:      emailHash,
		},
		Kayros: kayrosResp,
	}
}

// GenerateProofFilename creates a timestamped filename for the proof
func GenerateProofFilename(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	return fmt.Sprintf("kayros_proof_%s.json", t.Format("20060102_150405"))
}

// HandleKayrosBot processes incoming emails to the bot and sends a reply with Kayros proof
func HandleKayrosBot(req *IncomingEmailRequest) error {
	fmt.Println("--emailchain.HandleKayrosBot--")
	// Extract subject from email
	subject := ""
	subjectHeaders, err := extractHeaders(req.EmailRaw, []string{"Subject"})
	if err == nil && len(subjectHeaders) > 0 {
		subject = subjectHeaders[0]
	}

	// Hash the email
	emailHash := HashEmail(req.EmailRaw)
	fmt.Printf("Bot: Processing email from %v, hash: %s\n", req.From, emailHash)

	// Query Kayros API
	kayrosResp, err := QueryKayros(emailHash)
	if err != nil {
		fmt.Printf("Bot: Kayros API error: %v\n", err)
		return fmt.Errorf("kayros API query failed: %w", err)
	}

	// Create proof structure
	proof := CreateKayrosProof(req.From, req.To, subject, req.EmailRaw, req.Timestamp, kayrosResp)

	// Marshal proof to JSON
	proofJSON, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal proof: %w", err)
	}

	// Generate filename
	filename := GenerateProofFilename(req.Timestamp)

	// Build reply email with attachment
	err = SendKayrosBotReply(req.From, filename, proofJSON, req.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	fmt.Printf("Bot: Successfully sent reply to %v\n", req.From)
	return nil
}

// SendKayrosBotReply sends an email reply with the Kayros proof JSON attached
func SendKayrosBotReply(toAddresses []string, filename string, proofJSON []byte, timestamp int64) error {
	// Load DKIM signing options
	opts := LoadDkimKey()
	if opts == nil {
		return fmt.Errorf("no DKIM keys configured")
	}

	// Connect to database
	err := ConnectSql(ConnectionId)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Build email envelope
	from := vmimap.AddressFromString(BotEmail, "Kayros Bot")
	to := vmimap.AddressesFromString(toAddresses)

	date := time.Unix(timestamp, 0).UTC()
	subject := "Indexed by Kayros"

	envelope := &vmimap.Envelope{
		Subject: subject,
		From:    []vmimap.Address{from},
		To:      to,
		Date:    date,
	}

	// Build headers
	headers, err := SerializeEnvelope2(envelope, mail.Header{})
	if err != nil {
		return fmt.Errorf("failed to serialize envelope: %w", err)
	}

	// Email body text
	bodyText := fmt.Sprintf("Indexed by Kayros. See attached %s", filename)

	// Create email with attachment
	// Use multipart/mixed for attachment
	boundary := fmt.Sprintf("boundary_%d", timestamp)
	email := vmimap.Email{
		Headers: headers,
		Body: vmimap.EmailBody{
			ContentType: fmt.Sprintf("multipart/mixed; boundary=\"%s\"", boundary),
			Boundary:    boundary,
			Parts: []vmimap.BodyPart{
				{
					ContentType: "text/plain; charset=UTF-8",
					Body:        []byte(bodyText),
				},
			},
		},
		Attachments: []vmimap.Attachment{
			{
				Filename:    filename,
				ContentType: "application/json",
				Data:        proofJSON,
			},
		},
	}

	// Build raw email
	emailStr, err := BuildRawEmail2(email)
	if err != nil {
		return fmt.Errorf("failed to build email: %w", err)
	}

	// Prepare email for sending (add Message-ID and DKIM signature)
	prepped, err := prepareEmailSend(*opts, emailStr, BotEmail, date, true)
	if err != nil {
		return fmt.Errorf("failed to prepare email: %w", err)
	}

	// Send email to each recipient
	errs := []string{}
	for _, to := range toAddresses {
		err = sendEmailInternal(BotEmail, to, prepped, MailServerDomain, DefaultNetworkType)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", to, err))
			fmt.Printf("Bot: Failed to send to %s: %v\n", to, err)
			continue
		}
		fmt.Printf("Bot: Sent email to %s\n", to)
	}

	// Store sent email in bot's SENT folder
	err = StoreEmail(BotEmail, []string{}, []byte(prepped), "", ConnectionId, FolderSent)
	if err != nil {
		errs = append(errs, fmt.Sprintf("store error: %v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("send errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
