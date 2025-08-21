package wasmx

// SDK mirrors basic call helpers from the AS SDK.
type SDK struct {
    RoleOrAddress    Bech32String
    CallerModuleName string
}

func NewSDK(callerModuleName string) *SDK {
    return &SDK{CallerModuleName: callerModuleName}
}

func (s *SDK) Call(calld string, isQuery bool) CallResponse {
    // delegate to utils.CallContract for consistent behavior
    return CallContract(s.RoleOrAddress, calld, isQuery, s.CallerModuleName)
}

func (s *SDK) Query(calld string) CallResponse { return s.Call(calld, true) }
func (s *SDK) Execute(calld string) CallResponse { return s.Call(calld, false) }

func (s *SDK) callSafe(calld string, isQuery bool) string {
    resp := s.Call(calld, isQuery)
    if resp.Success > 0 {
        Revert([]byte("call to " + string(s.RoleOrAddress) + " failed: " + resp.Data))
    }
    return resp.Data
}

func (s *SDK) QuerySafe(calld string) string { return s.callSafe(calld, true) }
func (s *SDK) ExecuteSafe(calld string) string { return s.callSafe(calld, false) }

