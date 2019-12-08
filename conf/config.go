package conf

type ShortFormConfig struct {
	Secret string `json:"-"` // is never written to disk

	SecretEncoded string `json:"time"`
}
