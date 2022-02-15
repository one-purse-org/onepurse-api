package dal

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type IAdminDAL interface {
	AddRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, ID string) error
	AddAdmin(ctx context.Context, admin *model.Admin) error
	FindAdmin(ctx context.Context, query bson.D) (*model.Admin, error)
	FindAdmins(ctx context.Context, query bson.D) (*[]model.Admin, error)
	UpdateAdmin(ctx context.Context, ID string, updateParam bson.D) error
	DeleteAdmin(ctx context.Context, ID string) error
}

type AdminDAL struct {
	DB              *mongo.Database
	AdminCollection *mongo.Collection
	RoleCollection  *mongo.Collection
}

func NewAdminDAL(db *mongo.Database) *AdminDAL {
	return &AdminDAL{
		DB:              db,
		AdminCollection: db.Collection("admin"),
		RoleCollection:  db.Collection("admin_roles"),
	}
}

// AddRole ...
func (a AdminDAL) AddRole(ctx context.Context, role *model.Role) error {
	_, err := a.AdminCollection.InsertOne(ctx, role)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("roles already exists")
		}
		return err
	}
	return nil
}

// DeleteRole ...
func (a AdminDAL) DeleteRole(ctx context.Context, ID string) error {
	var role model.Role
	err := a.RoleCollection.FindOneAndDelete(ctx, bson.D{{"_id", ID}}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logrus.Errorf("error deleting role %s : role does not exist", ID)
			return errors.New("role record does not exist")
		}
	}
	return nil
}

// AddAdmin ...
func (a AdminDAL) AddAdmin(ctx context.Context, admin *model.Admin) error {
	_, err := a.AdminCollection.InsertOne(ctx, admin)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("admin record already exists")
		}
		return err
	}
	return nil
}

// FindAdmin ...
func (a AdminDAL) FindAdmin(ctx context.Context, query bson.D) (*model.Admin, error) {
	var admin model.Admin
	err := a.AdminCollection.FindOne(ctx, query).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			findErr := fmt.Sprintf("record for admin not found")
			return nil, errors.New(findErr)
		}
		return nil, err
	}
	return &admin, nil
}

// FindAdmins ...
func (a AdminDAL) FindAdmins(ctx context.Context, query bson.D) (*[]model.Admin, error) {
	var admins []model.Admin

	cursor, err := a.AdminCollection.Find(ctx, query)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &[]model.Admin{}, nil
		}
		logrus.Fatalf("[Mongo]: error fetching admins: %s", err.Error())
		return nil, err
	}

	if err = cursor.All(ctx, &admins); err != nil {
		logrus.Fatalf("[Mongo]: error parsing mongo document to admins model: %s", err.Error())
		return nil, err
	}

	return &admins, err
}

// UpdateAdmin ...
func (a AdminDAL) UpdateAdmin(ctx context.Context, ID string, updateParam bson.D) error {
	result, err := a.AdminCollection.UpdateByID(ctx, ID, updateParam)
	if err != nil {
		log.Fatalf("[Mongo]: error update ageint %s: %s", ID, err.Error())
		return err
	}
	if result.MatchedCount == 0 {
		logrus.Fatalf("[Mongo]: error updating admin %s: admin record not found", err.Error())
		return errors.New("agent record not found")
	}
	return nil
}

// DeleteAdmin ...
func (a AdminDAL) DeleteAdmin(ctx context.Context, ID string) error {
	var admin model.Admin
	err := a.AdminCollection.FindOneAndDelete(ctx, bson.D{{"_id", ID}}).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logrus.Errorf("error deleting admin %s : user record does not exist", ID)
			return errors.New("admin record does not exist")
		}
	}
	return nil
}
