package vmsmtp

import (
	"encoding/json"
	"fmt"
	"strings"

	gosmtp "github.com/emersion/go-smtp"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func ClientConnect(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SmtpConnectionRequest
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
		if conn.Info.ServerUrl == req.ServerUrl {
			err := conn.Client.Noop()
			if err == nil {
				return prepareResponse(rnh, response)
			} else {
				_ = closeConnection(vctx, conn, connId)
			}
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	getClient := func() (*gosmtp.Client, error) {
		c, err := connectSmtpClient(ctx.Context.GoContextParent, req.ServerUrl, req.StartTLS, req.NetworkType, req.Auth, req.TlsConfig)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return connectCommon(ctx, rnh, vctx, getClient, response, connId, req)
}

func connectCommon(
	ctx *Context,
	rnh memc.RuntimeHandler,
	vctx *SmtpContext,
	getClient func() (*gosmtp.Client, error),
	response *SmtpConnectionResponse,
	connId string,
	info SmtpConnectionRequest,
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
			err := client.Quit()
			if err != nil {
				err = client.Close()
			}
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`smtp close error for connection id "%s": %v`, connId, err))
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
		GoContextParent: ctx.GoContextParent,
		Info:            info,
		Client:          client,
		Closed:          closedChannel,
		GetClient:       getClient,
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	err = closeConnection(vctx, conn, connId)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func closeConnection(vctx *SmtpContext, conn *SmtpOpenConnection, connId string) (err error) {
	// for closing channel errors when already closed
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("closing connection: %v", r)
		}
	}()
	err = conn.Client.Quit()
	if err != nil {
		err = conn.Client.Close()
	}
	close(conn.Closed) // signal closing the database
	vctx.DeleteConnection(connId)
	return err
}

func Quit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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

func Hello(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	fmt.Println("====-Hello----")
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	fmt.Println("====-Hello----", string(requestbz))
	var req SmtpHelloRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SmtpHelloResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	err = conn.Client.Hello(req.LocalName)
	fmt.Println("==HELLO==", req.Id, "--", req.LocalName, err)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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
	fmt.Println("==SendMail==", req.Id)

	response := &SmtpSendMailResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "SMTP connection not found"
		return prepareResponse(rnh, response)
	}
	msgreader := strings.NewReader(string(req.Email))
	fmt.Println("--SendMail from,to--", req.From, req.To)

	fmt.Println("===============sendMail")
	fmt.Println(string(req.Email))
	fmt.Println("=====================")

	err = conn.Client.SendMail(req.From, req.To, msgreader)
	fmt.Println("--SendMail--", err)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func BuildMail(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
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

func ServerStart(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ServerStartResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	fmt.Println("--SMTP.ServerStart--", string(requestbz))
	var req ServerStartRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}
	if req.ConnectionId == "" {
		response.Error = "smtp server connection id empty"
		return prepareResponse(rnh, response)
	}
	connId := buildConnectionId(req.ConnectionId, ctx)
	_, ok := vctx.GetServerConnection(connId)
	if ok {
		response.Error = "already started"
		return prepareResponse(rnh, response)
	}
	s, err := NewServer(req.ServerConfig, ctx, req.ConnectionId)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	vctx.SetServerConnection(connId, &SmtpServerConnection{
		Server:          s,
		GoContextParent: ctx.GoContextParent,
		ContractAddress: ctx.ContractInfo.Address,
	})
	return prepareResponse(rnh, response)
}

func ServerClose(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ServerCloseResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ServerCloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}
	connId := buildConnectionId(req.ConnectionId, ctx)
	conn, ok := vctx.GetServerConnection(connId)
	if ok {
		vctx.DeleteServerConnection(connId)
		err = conn.Server.Close()
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
	}
	return prepareResponse(rnh, response)
}

func ServerShutdown(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &ServerCloseResponse{Error: ""}
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ServerShutdownRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	vctx, err := GetSmtpContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}
	connId := buildConnectionId(req.ConnectionId, ctx)
	conn, ok := vctx.GetServerConnection(connId)
	if ok {
		vctx.DeleteServerConnection(connId)
		err = conn.Server.Shutdown(ctx.Context.GoContextParent)
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
	}
	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

// per session
func buildConnectionId(id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s", ctx.Env.Contract.Address.String(), id)
}
