package deps

import (
	"github.com/isongjosiah/work/onepurse-api/config"
	userdal "github.com/isongjosiah/work/onepurse-api/dal"
	"github.com/isongjosiah/work/onepurse-api/services"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	// Services
	AWS *services.AWS
	// DAL
	DAL *userdal.DAL
}

func New(cfg *config.Config) (*Dependencies, error) {
	dal, err := userdal.New(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "[DEPS]: unable to set up DAL")
	}
	logrus.Info("[DAL]: OK")

	aws, err := services.NewAWS(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "[AWS]: unable to set up AWS services")
	}

	deps := &Dependencies{
		AWS: aws,
		DAL: dal,
	}

	return deps, nil
}
