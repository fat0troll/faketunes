package domains

type Domain interface {
	ConnectDependencies() error
	Start() error
}
