package types

const (
	EventTypeRegisterRoute         = "register_route"
	EventTypeDeregisterRoute       = "deregister_route"
	EventTypeRegisterOauthClient   = "register_outh_client"
	EventTypeDeregisterOauthClient = "deregister_outh_client"
	EventTypeEditOauthClient       = "edit_outh_client"
)

const (
	AttributeKeyRoute         = "route"
	AttributeKeyContract      = "contract_address"
	AttributeKeyOauthClientId = "oauth_client_id"
)
