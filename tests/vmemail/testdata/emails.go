package testdata

import (
	_ "embed"
)

var (
	//go:embed email/email_arc3.eml
	EmailARC3 string

	//go:embed email/email_arc1.eml
	Email1 string
)
