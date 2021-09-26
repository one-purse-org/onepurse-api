package dal

import (
	"go.mongodb.org/mongo-driver/mongo"
)
type DAL struct {
	DB *mongo.Database
}
