package external

type (
	IExampleExtService interface {
		GetExtData()
	}
	exampleExtService struct {
		url string
	}
)

func (p *pkgExternal) NewExampleExtService() IExampleExtService {
	return &exampleExtService{
		url: cfg.URL.ExampleExternalService,
	}
}

func (e *exampleExtService) GetExtData() {
	// TODO: implement me
}
