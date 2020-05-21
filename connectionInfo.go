package socketcommunication

type ConnectionType string

const (
	TCP  ConnectionType = "tcp"
	UNIX ConnectionType = "unix"
)

type ConnectionInfo struct {
	Ctype   ConnectionType
	Address string
}

func (ct ConnectionType) toString() string {
	return string(ct)
}
