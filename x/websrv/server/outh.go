package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"wasmx/v1/x/websrv/types"
)

type SignMessage struct {
	Scope string `json:"scope"`
	Nonce string `json:"nonce"`
}

type Coin struct {
	Amount string `json:"amount"`
	Denom  string `json:"denom"`
}

type StdFee struct {
	Amount []Coin `json:"amount"`
	Gas    string `json:"gas"`
}

type MsgSignData struct {
	Data   string `json:"data"` // base64
	Signer string `json:"signer"`
}

type AminoMsg struct {
	Type  string      `json:"type"`
	Value MsgSignData `json:"value"`
}

type SignDocAmino struct {
	AccountNumber string     `json:"account_number"`
	ChainId       string     `json:"chain_id"`
	Fee           StdFee     `json:"fee"`
	Memo          string     `json:"memo"`
	Msgs          []AminoMsg `json:"msgs"`
	Sequence      string     `json:"sequence"`
}

var (
// //go:embed static/login.html
// loginFS embed.FS

// //go:embed static/auth.html
// authFS embed.FS
)

var AUTHORIZATION_CODE_EXPIRATION = time.Minute * 5

// User:rwx Group:r-x World:---
var secretPerm fs.FileMode = 0o750

// temporary nonces
var nonceDB = map[string]time.Time{}

var (
	tokenDbName = "tokendb"
	dumpvar     bool
)

func (k WebsrvServer) InitOauth2(mux *http.ServeMux, dirname string) {
	tokenDbPath := path.Join(dirname, tokenDbName)
	dumpvar = false

	if err := os.MkdirAll(dirname, secretPerm); err != nil {
		panic(err)
	}

	flag.Parse()
	if dumpvar {
		log.Println("Dumping requests")
	}
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// token store
	manager.MustTokenStorage(store.NewFileTokenStore(tokenDbPath))

	// generate jwt access token
	// manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	// var clients []types.OauthClientInfo

	// clientsResp, _ := k.queryClient.GetAllOauthClients(k.ctx, &types.QueryGetAllOauthClientsRequest{})
	// if clientsResp != nil {
	// 	clients = clientsResp.Clients
	// }
	clientStore := store.NewClientStore()
	// for _, client := range clients {
	// 	id := types.OauthClientIdToString(client.ClientId)
	// 	clientStore.Set(id, &models.Client{
	// 		ID:     id,
	// 		Secret: "",
	// 		Domain: client.Domain,
	// 		Public: true,
	// 	})
	// }

	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	// TODO remove this
	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
		if username == "test" && password == "test" {
			userID = "test"
		}
		return
	})

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err)
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error)
	})

	mux.HandleFunc("/login", k.RouteOAuthLogin)
	mux.HandleFunc("/auth", k.RouteOAuthAuth)
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		if dumpvar {
			dumpRequest(os.Stdout, "authorize", r)
		}

		clientIdHex := r.URL.Query().Get("client_id")
		clientId, err := types.OauthClientIdFromString(clientIdHex)
		if err == nil {
			client, err := k.queryClient.GetOauthClient(k.ctx, &types.QueryGetOauthClientRequest{ClientId: clientId})
			if err == nil && client != nil && client.Client != nil {
				clientStore.Set(clientIdHex, &models.Client{
					ID:     clientIdHex,
					Secret: "",
					Domain: client.Client.Domain,
					Public: true,
				})
			}
		}

		store, err := session.Start(r.Context(), w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var form url.Values
		if v, ok := store.Get("ReturnUri"); ok {
			form = v.(url.Values)
		}
		r.Form = form

		store.Delete("ReturnUri")
		store.Save()

		err = srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if dumpvar {
			_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
		}

		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if dumpvar {
			_ = dumpRequest(os.Stdout, "test", r) // Ignore the error
		}
		token, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data := map[string]interface{}{
			"expires_in": int64(time.Until(
				token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()),
			).Seconds()),
			"client_id": token.GetClientID(),
			"user_id":   token.GetUserID(),
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(data)
	})
}

func (k WebsrvServer) RouteOAuthLogin(w http.ResponseWriter, r *http.Request) {
	chainId := k.longChainID

	if dumpvar {
		_ = dumpRequest(os.Stdout, "login", r) // Ignore the error
	}
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		messageStr := r.Form.Get("message")
		pubkeyHex := r.Form.Get("pubkey")
		// algo := r.Form.Get("algo")
		signatureEncoded := r.Form.Get("signature")

		pubKeyBz, err := hex.DecodeString(pubkeyHex)
		if err != nil {
			http.Error(w, "Failed to url decode pubKey from hex", http.StatusInternalServerError)
			return
		}
		pubKey := secp256k1.PubKey{Key: pubKeyBz}
		signature, err := base64.StdEncoding.DecodeString(signatureEncoded)
		if err != nil {
			http.Error(w, "Failed to base64 decode signature", http.StatusInternalServerError)
			return
		}
		bech32address := sdk.AccAddress(pubKey.Address())

		msgSigned, err := getStdSignDoc([]byte(messageStr), bech32address.String())
		if err != nil {
			http.Error(w, "Failed to build StdSignDoc", http.StatusInternalServerError)
			return
		}

		verified := pubKey.VerifySignature(msgSigned, signature)
		if !verified {
			http.Error(w, "Failed to verify signature", http.StatusUnauthorized)
			return
		}

		var message SignMessage
		err = json.Unmarshal([]byte(messageStr), &message)
		if err != nil {
			http.Error(w, "Failed to parse signed message", http.StatusInternalServerError)
			return
		}

		expiration, found := nonceDB[message.Nonce]
		if !found {
			http.Error(w, "invalid nonce", http.StatusUnauthorized)
			return
		}
		if expiration.Unix() < time.Now().Unix() {
			delete(nonceDB, message.Nonce)
			http.Error(w, "signed nonce expired", http.StatusUnauthorized)
			return
		}

		// TODO what to do with the nonce?
		delete(nonceDB, message.Nonce)

		store.Set("LoggedInUserID", bech32address.String())
		store.Save()

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}

	// Create a random nonce to prevent replay attacks
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		http.Error(w, "Failed to generate nonce", http.StatusInternalServerError)
		return
	}

	nonceExpiration := time.Now().Add(AUTHORIZATION_CODE_EXPIRATION)

	// Encode the nonce as a base64 string
	nonceStr := hex.EncodeToString(nonce)
	nonceDB[nonceStr] = nonceExpiration

	// Define the message to be signed
	message := SignMessage{
		Scope: "profile",
		Nonce: nonceStr,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Failed to parse SignMessage", http.StatusInternalServerError)
		return
	}

	// Render the authorization page that prompts the user to sign the message with Keplr
	// TODO fix relative path somehow
	// tmpl, err := template.ParseFS(loginFS)
	tmpl, err := template.ParseFiles("./x/websrv/server/static/login.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, struct {
		Message          string
		AuthorizationURL string
		ChainId          string
	}{
		Message: string(messageJson),
		ChainId: chainId,
	}); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (k WebsrvServer) RouteOAuthAuth(w http.ResponseWriter, r *http.Request) {
	if dumpvar {
		_ = dumpRequest(os.Stdout, "auth", r) // Ignore the error
	}
	store, err := session.Start(k.ctx, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	outputHTML(w, r, "./x/websrv/server/static/auth.html")
	// http.FileServer(http.FS(authFS))
}

func getStdSignDoc(msg []byte, bech32address string) ([]byte, error) {
	messageStrBase64 := base64.StdEncoding.EncodeToString(msg)
	msgDocAmino := SignDocAmino{
		AccountNumber: "0",
		Sequence:      "0",
		ChainId:       "",
		Fee:           StdFee{Gas: "0", Amount: []Coin{}},
		Memo:          "",
		Msgs:          []AminoMsg{{Type: "sign/MsgSignData", Value: MsgSignData{Signer: bech32address, Data: messageStrBase64}}},
	}
	msgSigned, err := json.Marshal(msgDocAmino)
	if err != nil {
		return nil, err
	}
	return msgSigned, nil
}

func dumpRequest(writer io.Writer, header string, r *http.Request) error {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}
	writer.Write([]byte("\n" + header + ": \n"))
	writer.Write(data)
	return nil
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	if dumpvar {
		_ = dumpRequest(os.Stdout, "userAuthorizeHandler", r) // Ignore the error
	}
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return
	}

	uid, ok := store.Get("LoggedInUserID")
	if !ok {
		if r.Form == nil {
			r.ParseForm()
		}

		store.Set("ReturnUri", r.Form)
		store.Save()

		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	userID = uid.(string)
	store.Delete("LoggedInUserID")
	store.Save()
	return
}
