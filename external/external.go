package external

import "github.com/golang-base-template/util/config"

var (
	cfg = config.Config{}
)

type (
	IPkgExternal interface {
		NewExampleExtService() IExampleExtService
	}
	pkgExternal struct {
		cfg config.Config
	}
)

func NewPkgExternal() IPkgExternal {
	return &pkgExternal{
		cfg: config.Get(),
	}
}
