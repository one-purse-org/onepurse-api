package dal

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ICurrencyDAL interface {
	Add(ctx context.Context, currency *model.Currency) error
	FindAll(ctx context.Context) (*[]model.Currency, error)
}

type CurrencyDAL struct {
	DB         *mongo.Database
	Collection *mongo.Collection
}

func NewCurrencyDAL(db *mongo.Database) *CurrencyDAL {
	return &CurrencyDAL{
		DB:         db,
		Collection: db.Collection("currency"),
	}
}

func (c CurrencyDAL) Add(ctx context.Context, currency *model.Currency) error {
	_, err := c.Collection.InsertOne(ctx, currency)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("BaseCurrency already exists")
		} else {
			return err
		}
	}
	return nil
}

func (c CurrencyDAL) FindAll(ctx context.Context) (*[]model.Currency, error) {
	var currency *[]model.Currency
	cursor, err := c.Collection.Find(ctx, bson.D{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.Currency{}, nil
		} else {
			logrus.Fatalf("[Mongo]: error fetching currencies: %s", err.Error())
			return nil, err
		}
	}

	if err = cursor.All(ctx, &currency); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to currency model: %s", err.Error())
		return nil, err
	}

	return currency, nil
}
