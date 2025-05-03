package templates

import "embed"

//go:embed email.html
var EmailTemplateFS embed.FS
