package model
type CoConfig struct {
	Server struct {
		Port string
		Metricsport string
	}
	Appregistry struct {
		Repo string
		Branch string
	}
	Git struct {
		Repo string
	}
}
