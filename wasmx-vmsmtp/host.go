package vmsmtp

import (
	"encoding/json"
	"fmt"
	"strings"

	gosmtp "github.com/emersion/go-smtp"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ConnectWithPassword(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpConnectionSimpleRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpConnectionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.SmtpServerUrlSTARTTLS == req.SmtpServerUrlSTARTTLS || conn.SmtpServerUrlTLS == req.SmtpServerUrlTLS {
			return prepareResponse(rnh, response)
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	getClient := func() (*gosmtp.Client, error) {
		c, err := connectToSMTP(req.SmtpServerUrlSTARTTLS, req.SmtpServerUrlTLS, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return connectCommon(ctx, rnh, vctx, getClient, response, connId, req.Username, req.SmtpServerUrlSTARTTLS, req.SmtpServerUrlTLS)
}

func ConnectOAuth2(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpConnectionOauth2Request
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpConnectionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.SmtpServerUrlSTARTTLS == req.SmtpServerUrlSTARTTLS || conn.SmtpServerUrlTLS == req.SmtpServerUrlTLS {
			return prepareResponse(rnh, response)
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	getClient := func() (*gosmtp.Client, error) {
		c, err := connectToSMTPOauth2(req.SmtpServerUrlSTARTTLS, req.SmtpServerUrlTLS, req.Username, req.AccessToken)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return connectCommon(ctx, rnh, vctx, getClient, response, connId, req.Username, req.SmtpServerUrlSTARTTLS, req.SmtpServerUrlTLS)
}

func connectCommon(
	ctx *Context,
	rnh memc.RuntimeHandler,
	vctx *SmtpContext,
	getClient func() (*gosmtp.Client, error),
	response *SmtpConnectionResponse,
	connId string,
	username string,
	imapServerUrlStartTls string,
	imapServerUrlTls string,
) ([]interface{}, error) {
	client, err := getClient()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	closedChannel := make(chan struct{})

	// TODO this should be done in 1 cleanup hook per vm extension
	ctx.GoRoutineGroup.Go(func() error {
		select {
		case <-ctx.GoContextParent.Done():
			ctx.Ctx.Logger().Info(fmt.Sprintf("parent context was closed, closing database connection: %s", connId))
			err := client.Close()
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`database close error for connection id "%s": %v`, connId, err))
			}
			close(closedChannel)
			return nil
		case <-closedChannel:
			// when close signal is received from Close() API
			// database is already closed, so we exit this goroutine
			ctx.Ctx.Logger().Info(fmt.Sprintf("database connection closed: %s", connId))
			return nil
		}
	})

	conn := &SmtpOpenConnection{
		GoContextParent:       ctx.GoContextParent,
		Username:              username,
		SmtpServerUrlSTARTTLS: imapServerUrlStartTls,
		SmtpServerUrlTLS:      imapServerUrlTls,
		Client:                client,
		Closed:                closedChannel,
		GetClient:             getClient,
	}

	err = vctx.SetConnection(connId, conn)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Close(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpCloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpCloseResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	err = conn.Client.Close()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	close(conn.Closed) // signal closing the database
	return prepareResponse(rnh, response)
}

func Quit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpQuitRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpQuitResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	err = conn.Client.Quit()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	close(conn.Closed) // signal closing the database
	return prepareResponse(rnh, response)
}

func Extension(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpExtensionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpExtensionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	exists, value := conn.Client.Extension(req.Name)
	response.Found = exists
	response.Params = value
	return prepareResponse(rnh, response)
}

func Noop(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpNoopRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpNoopResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	err = conn.Client.Noop()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpVerifyRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpVerifyResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	err = conn.Client.Verify(req.Address)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func SupportsAuth(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpSupportsAuthRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpSupportsAuthResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	exists := conn.Client.SupportsAuth(req.Mechanism)
	response.Found = exists
	return prepareResponse(rnh, response)
}

func MaxMessageSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpMaxMessageSizeRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpMaxMessageSizeResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	size, ok := conn.Client.MaxMessageSize()
	response.Size = int64(size)
	response.Ok = ok
	return prepareResponse(rnh, response)
}

func SendMail(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpSendMailRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpSendMailResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	msgreader := strings.NewReader(string(req.Email))
	err = conn.Client.SendMail(req.From, req.To, msgreader)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func BuildMail(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req SmtpBuildMailRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	response := &SmtpBuildMailResponse{Error: ""}
	raw, err := BuildRawEmail(req.Email)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Data = []byte(raw)
	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

// per session
func buildConnectionId(id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s", ctx.Env.Contract.Address.String(), id)
}

func BuildWasmxSmtpVM(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("ConnectWithPassword", ConnectWithPassword, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ConnectOAuth2", ConnectOAuth2, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Quit", Quit, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Extension", Extension, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Noop", Noop, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SendMail", SendMail, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Verify", Verify, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("SupportsAuth", SupportsAuth, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("MaxMessageSize", MaxMessageSize, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("BuildMail", BuildMail, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "smtp", context, fndefs)
}
