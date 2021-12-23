package dal

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IAgentDAL interface {
	Add(agent *model.Agent) error
	FindAll(query bson.D) (*[]model.Agent, error)
	FindOne(query bson.D) (*model.Agent, error)
}

type AgentDAL struct {
	DB         *mongo.Database
	Collection *mongo.Collection
}

func NewAgentDAL(db *mongo.Database) *AgentDAL {
	return &AgentDAL{
		DB:         db,
		Collection: db.Collection("agent"),
	}
}

func (a AgentDAL) Add(agent *model.Agent) error {
	_, err := a.Collection.InsertOne(context.TODO(), agent)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("agent record already exists")
		}
		return err
	}
	return nil
}

func (a AgentDAL) FindOne(query bson.D) (*model.Agent, error) {
	var agent model.Agent
	err := a.Collection.FindOne(context.TODO(), query).Decode(&agent)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for agent not found")
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return &agent, nil
}

func (a AgentDAL) FindAll(query bson.D) (*[]model.Agent, error) {
	var agents []model.Agent

	cursor, err := a.Collection.Find(context.TODO(), query)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.Agent{}, nil
		}
		logrus.Fatalf("[Mongo]: error fetching agents: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(context.TODO(), &agents); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to agents model: %s", err.Error())
		return nil, err
	}
	return &agents, err
}
