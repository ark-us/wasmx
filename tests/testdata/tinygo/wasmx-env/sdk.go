package wasmx

type SDK struct {
	RoleOrAddress       Bech32String
	CallerModuleName    string
	Revert              func(message string)
	LoggerInfo          func(msg string, parts []string)
	LoggerError         func(msg string, parts []string)
	LoggerDebug         func(msg string, parts []string)
	LoggerDebugExtended func(msg string, parts []string)
}

func NewSDK(
	callerModuleName string,
	revert func(message string),
	loggerInfo func(msg string, parts []string),
	loggerError func(msg string, parts []string),
	loggerDebug func(msg string, parts []string),
	loggerDebugExtended func(msg string, parts []string),
) *SDK {
	return &SDK{
		CallerModuleName:    callerModuleName,
		Revert:              revert,
		LoggerInfo:          loggerInfo,
		LoggerError:         loggerError,
		LoggerDebug:         loggerDebug,
		LoggerDebugExtended: loggerDebugExtended,
	}
}

func (s *SDK) Query(calld string) (bool, []byte) {
	return CallStatic(s.RoleOrAddress, []byte(calld), bigInt(DEFAULT_GAS_TX), s.CallerModuleName)
}
func (s *SDK) Execute(calld string) (bool, []byte) {
	return Call(s.RoleOrAddress, nil, []byte(calld), bigInt(DEFAULT_GAS_TX), s.CallerModuleName)
}

func (s *SDK) CallSafe(calld string, isQuery bool) []byte {
	ok, data := CallInternal(s.RoleOrAddress, nil, []byte(calld), bigInt(DEFAULT_GAS_TX), isQuery, s.CallerModuleName)
	if !ok {
		s.Revert("call to " + string(s.RoleOrAddress) + " failed: " + string(data))
	}
	return data
}

func (s *SDK) QuerySafe(calld string) []byte   { return s.CallSafe(calld, true) }
func (s *SDK) ExecuteSafe(calld string) []byte { return s.CallSafe(calld, false) }
