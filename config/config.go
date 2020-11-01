package config

// Interface representing interactions with short-form config files.
type Config interface {
	GetDatabasePath() string
	SetDatabasePath(path string) error
	Save() error
}
