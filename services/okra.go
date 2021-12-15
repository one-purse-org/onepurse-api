package services

import (
	okra "github.com/Uchencho/OkraGo"
	"github.com/isongjosiah/work/onepurse-api/config"
)

type OKRA struct {
	config      *config.Config
	okraClilent *okra.Client
}

func NewOkraService(cfg *config.Config) (*OKRA, error) {
	okraClient, err := okra.New(cfg.OkraToken, "https://api.okra.ng/snadbox/v1/")
	if err != nil {
		return nil, err
	}

	o := OKRA{
		config:      cfg,
		okraClilent: &okraClient,
	}
	return &o, nil
}
