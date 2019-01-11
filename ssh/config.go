package ssh

type Server struct {
	Hostname   string
	Port       int
	User       string
	Password   string
	PrivateKey string
}
