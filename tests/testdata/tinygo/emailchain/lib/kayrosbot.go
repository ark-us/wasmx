package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	vmimap "github.com/loredanacirstea/wasmx-env-imap/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	verifier "github.com/loredanacirstea/wasmx-kayros-verifier/lib"
)

const (
	BotEmail      = "kayros1@dmail.provable.dev"
	KayrosAPIURL  = "https://kayros.provable.dev"
	KayrosAPIGET  = KayrosAPIURL + "/api/database/record-by-hash?hash_item="
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
	Data   string      `json:"data"`
	Kayros interface{} `json:"kayros"`
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
	dataItem, err := hex.DecodeString(emailHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode email hash: %w", err)
	}
	client := verifier.NewKayrosClient(verifier.KayrosConfig{
		ApiBaseUrl: KayrosAPIURL,
		ApiUserKey: "",
	})
	regResp, err := client.RegisterData(verifier.KayrosRegistrationRequest{
		DataType: DataTypeEmail,
		DataItem: dataItem,
	})
	if err != nil {
		return nil, fmt.Errorf("Kayros registration failed: %w", err)
	}
	// Keep backward-compatible shape expected by the bot reply flow.
	return map[string]interface{}{
		"success": regResp.Success,
		"data": map[string]interface{}{
			"computed_hash_hex": regResp.Hash,
		},
		"timeuuid": regResp.TimeUUID,
		"encoding": regResp.Encoding,
	}, nil
}

// CreateKayrosProof creates a proof structure combining email data and Kayros response
func CreateKayrosProof(emailRaw []byte, kayrosResp interface{}) *KayrosProof {
	return &KayrosProof{
		Data:   base64.StdEncoding.EncodeToString(emailRaw),
		Kayros: kayrosResp,
	}
}

// GenerateProofFilename creates a timestamped filename for the proof
func GenerateProofFilename(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	return fmt.Sprintf("kayros_proof_%s.json", t.Format("20060102_150405"))
}

// unfoldHeader removes folding whitespace from email headers
// Email headers can be folded across multiple lines using \r\n followed by whitespace
func unfoldHeader(value string) string {
	// Replace \r\n followed by whitespace with a single space
	value = strings.ReplaceAll(value, "\r\n ", " ")
	value = strings.ReplaceAll(value, "\r\n\t", " ")
	// Also handle just \n (some implementations are sloppy)
	value = strings.ReplaceAll(value, "\n ", " ")
	value = strings.ReplaceAll(value, "\n\t", " ")
	// Remove any remaining \r\n
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.TrimSpace(value)
}

// HandleKayrosBot processes incoming emails to the bot and sends a reply with Kayros proof
func HandleKayrosBot(req *IncomingEmailRequest) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("HandleKayrosBot panic: %v", r)
		}
	}()

	wasmx.LoggerInfo(MODULE_NAME, "kayrosbot handler", []string{})
	// Extract subject, Message-ID, and References from email
	subject := "Indexed by Kayros"
	var messageID, references string

	headerValues, err := extractHeaders(req.EmailRaw, []string{"Subject", "Message-ID", "References"})
	if err == nil && len(headerValues) >= 1 {
		if len(headerValues[0]) > 0 {
			subject = unfoldHeader(headerValues[0])
		}
		if len(headerValues) >= 2 && len(headerValues[1]) > 0 {
			messageID = unfoldHeader(headerValues[1])
		}
		if len(headerValues) >= 3 && len(headerValues[2]) > 0 {
			references = unfoldHeader(headerValues[2])
		}
	}

	// Hash the email
	emailHash := HashEmail(req.EmailRaw)
	wasmx.LoggerInfo(MODULE_NAME, "processing email", []string{"from", req.From[0], "hash", emailHash})

	// Query Kayros API
	kayrosResp, err := QueryKayros(emailHash)
	if err != nil {
		wasmx.LoggerError(MODULE_NAME, "Kayros API error", []string{"error", err.Error()})
		return fmt.Errorf("kayros API query failed: %w", err)
	}

	// Extract computed_hash_hex from Kayros response
	computedHash := ""
	if kayrosData, ok := kayrosResp.(map[string]interface{}); ok {
		if data, ok := kayrosData["data"].(map[string]interface{}); ok {
			if hash, ok := data["computed_hash_hex"].(string); ok {
				computedHash = hash
			}
		}
	}

	// Create proof structure
	proof := CreateKayrosProof(req.EmailRaw, kayrosResp)

	// Marshal proof to JSON
	proofJSON, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal proof: %w", err)
	}

	// Generate filename
	filename := GenerateProofFilename(req.Timestamp)

	// Build reply email with attachment
	err = SendKayrosBotReply(req.From, filename, proofJSON, req.Timestamp, subject, messageID, references, computedHash)
	if err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}
	wasmx.LoggerInfo(MODULE_NAME, "sent reply to", []string{"from", req.From[0], "hash", emailHash})
	return nil
}

// SendKayrosBotReply sends an email reply with the Kayros proof JSON attached
func SendKayrosBotReply(toAddresses []string, filename string, proofJSON []byte, timestamp int64, subject string, messageID string, references string, computedHash string) error {
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
	from := vmimap.AddressFromString(BotEmail, "Kayros Indexer from provable.dev")
	to := vmimap.AddressesFromString(toAddresses)

	date := time.Unix(timestamp, 0).UTC()

	envelope := &vmimap.Envelope{
		Subject:   "Fwd: " + subject,
		From:      []vmimap.Address{from},
		To:        to,
		Date:      date,
		InReplyTo: []string{messageID},
	}

	// Build headers
	hdr := mail.Header{}

	// Build References header: original References + original Message-ID
	var replyReferences []string
	if references != "" {
		replyReferences = append(replyReferences, references)
	}
	if messageID != "" {
		replyReferences = append(replyReferences, messageID)
	}
	if len(replyReferences) > 0 {
		hdr.Set("References", strings.Join(replyReferences, " "))
	}

	headers, err := SerializeEnvelope2(envelope, hdr)
	if err != nil {
		return fmt.Errorf("failed to serialize envelope: %w", err)
	}

	// Email body text
	var bodyText string
	if computedHash != "" {
		bodyText = fmt.Sprintf(`Indexed by Kayros. See attached %s

View stored record on Kayros: %s%s
`, filename, KayrosAPIGET, computedHash)
	} else {
		bodyText = fmt.Sprintf(`Failure to create proof. See attached %s for details.
`, filename)
	}

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
			wasmx.LoggerError(MODULE_NAME, "failed to send", []string{"to", to, "error", err.Error()})
			continue
		}
		wasmx.LoggerInfo(MODULE_NAME, "sent email to", []string{"to", to})
	}

	// Store sent email in bot's SENT folder (unless disabled)
	if !LoadKayrosBotNoStore() {
		err = StoreEmail(BotEmail, []string{}, []byte(prepped), "", ConnectionId, FolderSent)
		if err != nil {
			errs = append(errs, fmt.Sprintf("store error: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("send errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
