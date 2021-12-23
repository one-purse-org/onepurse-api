package dal

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DAL struct {
	DB *mongo.Database
	// define a session to utilize transactions
	Session mongo.Session

	UserDAL        IUserDAL
	CurrencyDAL    ICurrencyDAL
	TransactionDAL ITransactionDAL
	AgentDAL       IAgentDAL
}

func (d *DAL) setupDALObjects(cfg *config.Config) error {
	// set up database
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return errors.Wrapf(err, "[Mongo]: unable to open intial connection")
	}
	if cfg.Environment == "development" {
		d.DB = client.Database("onepurse-dev")
	} else if cfg.Environment == "production" {
		d.DB = client.Database("onepurse")
	}

	d.Session, err = client.StartSession()
	if err != nil {
		return errors.Wrapf(err, "[Mongo]: unable to create a session")
	}
	d.UserDAL = NewUserDAL(d.DB)
	d.CurrencyDAL = NewCurrencyDAL(d.DB)
	d.TransactionDAL = NewTransactionDAL(d.DB)
	d.AgentDAL = NewAgentDAL(d.DB)
	return nil
}

func New(cfg *config.Config) (*DAL, error) {
	dal := &DAL{}

	if err := dal.setupDALObjects(cfg); err != nil {
		return nil, err
	}
	return dal, nil
}
