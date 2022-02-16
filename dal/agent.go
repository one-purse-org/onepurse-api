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
	Add(ctx context.Context, agent *model.Agent) error
	FindAll(ctx context.Context, query bson.D) (*[]model.Agent, error)
	FindOne(ctx context.Context, query bson.D) (*model.Agent, error)
	Update(ctx context.Context, agentID string, updateParam bson.D) error
	Count(ctx context.Context) (int32, error)
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

func (a AgentDAL) Add(ctx context.Context, agent *model.Agent) error {
	_, err := a.Collection.InsertOne(ctx, agent)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("agent record already exists")
		}
		return err
	}
	return nil
}

func (a AgentDAL) FindOne(ctx context.Context, query bson.D) (*model.Agent, error) {
	var agent model.Agent
	err := a.Collection.FindOne(ctx, query).Decode(&agent)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for agent not found")
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return &agent, nil
}

func (a AgentDAL) FindAll(ctx context.Context, query bson.D) (*[]model.Agent, error) {
	var agents []model.Agent

	cursor, err := a.Collection.Find(ctx, query)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.Agent{}, nil
		}
		logrus.Fatalf("[Mongo]: error fetching agents: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &agents); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to agents model: %s", err.Error())
		return nil, err
	}
	return &agents, err
}

func (a AgentDAL) Update(ctx context.Context, agentID string, updateParam bson.D) error {
	result, err := a.Collection.UpdateByID(ctx, agentID, updateParam)
	if err != nil {
		logrus.Errorf("[Mongo]: error update ageint %s: %s", agentID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Errorf("[Mongo]: error updating agent %s: agent record not found", err.Error())
		return errors.New("agent record not found")
	}
	return nil
}

func (a AgentDAL) Count(ctx context.Context) (int32, error) {
	num, err := a.Collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}
	return int32(num), err
}
