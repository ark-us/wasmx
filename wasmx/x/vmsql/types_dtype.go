package vmsql

var DbName = "dtype"
var DTypeConnection = "dtype_connection"
var DTypeDbName = "dtype_db"
var DTypeDbConnName = "dtype_db_connection"
var DTypeTableName = "dtype_table"
var DTypeFieldName = "dtype_field"

var DTypeNodeName = "node"
var DTypeRelationName = "relation"
var DTypeRelationTypeName = "relation_type"

var TokensTable = "token"
var OwnedTable = "owned"
var AllowanceTable = "allowance"

var IdentityTable = "identity"
var FullNameTable = "identity_full_name"
var EmailTable = "identity_email"

var TableDbConnId = int64(1)
var TableDbId = int64(2)
var TableTableId = int64(3)
var TableFieldsId = int64(4)
var TableNodeId = int64(5)
var TableRelationId = int64(6)
var TableRelationTypeId = int64(7)
var TokensTableId = int64(8)
var OwnedTableId = int64(9)
var AllowanceTableId = int64(10)
var IdentityTableId = int64(11)
var FullNameTableId = int64(12)
var EmailTableId = int64(13)

type InstantiateDType struct {
	Dir    string `json:"dir"`
	Driver string `json:"driver"`
}

type CreateTableDTypeRequest struct {
	TableId int64 `json:"table_id"`
}

type CreateTableDTypeResponse struct {
}

type TableIdentifier struct {
	DbConnectionId   int64  `json:"db_connection_id"`
	DbId             int64  `json:"db_id"`
	TableId          int64  `json:"table_id"`
	DbConnectionName string `json:"db_connection_name"`
	DbName           string `json:"db_name"`
	TableName        string `json:"table_name"`
}

type InsertDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Data       []byte          `json:"data"`
}

type UpdateDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Condition  []byte          `json:"condition"`
	Data       []byte          `json:"data"`
}

type DeleteDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Condition  []byte          `json:"condition"`
}

type ReadDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Data       []byte          `json:"data"`
}

type ReadFieldRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	FieldId    int64           `json:"fieldId"`
	FieldName  string          `json:"fieldName"`
	Data       []byte          `json:"data"`
}

type BuildSchemaRequest struct {
	Identifier TableIdentifier `json:"identifier"`
}

type BuildSchemaResponse struct {
	Data []byte `json:"data"`
}

type InsertDTypeResponse struct {
}

type ConnectRequest struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type InitializeTokens struct{}
type InitializeIdentity struct{}

type CalldataDType struct {
	Initialize         *InstantiateDType        `json:"Initialize,omitempty"`
	InitializeTokens   *InitializeTokens        `json:"InitializeTokens,omitempty"`
	InitializeIdentity *InitializeIdentity      `json:"InitializeIdentity,omitempty"`
	CreateTable        *CreateTableDTypeRequest `json:"CreateTable,omitempty"`
	Connect            *ConnectRequest          `json:"Connect,omitempty"`
	Close              *ConnectRequest          `json:"Close,omitempty"`
	Insert             *InsertDTypeRequest      `json:"Insert,omitempty"`
	InsertOrReplace    *InsertDTypeRequest      `json:"InsertOrReplace,omitempty"`
	Update             *UpdateDTypeRequest      `json:"Update,omitempty"`
	Delete             *DeleteDTypeRequest      `json:"Delete,omitempty"`
	Read               *ReadDTypeRequest        `json:"Read,omitempty"`
	ReadField          *ReadFieldRequest        `json:"ReadField,omitempty"`
	BuildSchema        *BuildSchemaRequest      `json:"BuildSchema,omitempty"`
}
