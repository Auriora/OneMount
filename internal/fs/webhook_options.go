package fs

import "time"

// WebhookOptions controls webhook subscription behaviour.
type WebhookOptions struct {
	Enabled          bool
	UseSocketIO      bool
	PublicURL        string
	ListenAddress    string
	Path             string
	ClientState      string
	TLSCertFile      string
	TLSKeyFile       string
	Resource         string
	ChangeType       string
	FallbackInterval time.Duration
}

// ConfigureWebhooks stores the webhook options for the filesystem. Must be invoked before DeltaLoop starts.
func (f *Filesystem) ConfigureWebhooks(opts WebhookOptions) {
	f.webhookOptions = &opts
}
