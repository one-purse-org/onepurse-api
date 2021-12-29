package dal

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DAL struct {
	Client          *mongo.Client
	DB              *mongo.Database
	UserDAL         IUserDAL
	CurrencyDAL     ICurrencyDAL
	TransactionDAL  ITransactionDAL
	AgentDAL        IAgentDAL
	NotificationDAL INotificationDAL
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

	d.Client = client
	d.UserDAL = NewUserDAL(d.DB)
	d.CurrencyDAL = NewCurrencyDAL(d.DB)
	d.TransactionDAL = NewTransactionDAL(d.DB)
	d.AgentDAL = NewAgentDAL(d.DB)
	d.NotificationDAL = NewNotificationDAL(d.DB)
	return nil
}

func New(cfg *config.Config) (*DAL, error) {
	dal := &DAL{}

	if err := dal.setupDALObjects(cfg); err != nil {
		return nil, err
	}
	return dal, nil
}
