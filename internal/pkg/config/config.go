package config

type Config struct {
	Certfile *string
	JSONFile *bool
	HTTPS    *bool
	Keyfile  *string
	Port     *uint
}
