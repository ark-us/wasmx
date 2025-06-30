package testdata

import (
	_ "embed"
)

var (
	//go:embed email/email_arc3.eml
	EmailARC3 string

	//go:embed email/email_arc1.eml
	Email1 string

	//go:embed email/email_dkim1.eml
	EmailDkim1 string

	//go:embed email/email_dkim2.eml
	EmailDkim2 string

	//go:embed email/email_forwarded_0.eml
	EmailForwarded0 string

	//go:embed email/email_forwarded_1.eml
	EmailForwarded1 string
)
