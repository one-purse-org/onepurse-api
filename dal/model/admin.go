package model

type Currency struct {
	Label string `bson:"label" json:"label"`
	Slug  string `bson:"slug" json:"slug"`
}
