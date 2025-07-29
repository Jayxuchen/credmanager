package types

import "time"

type Credential struct {
	Key      string
	Value    string
	Expiry   *time.Time
	Metadata map[string]string
}
