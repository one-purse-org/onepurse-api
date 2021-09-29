package deps

import (
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/dal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	// DAL
	DAL *dal.DAL
}

func New(cfg *config.Config) (*Dependencies, error) {
	dal, err := dal.New(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "[DEPS]: unable to set up DAL")
	}
	logrus.Info("[DAL]: OK")

	deps := &Dependencies{
		DAL: dal,
	}

	return deps, nil
}
