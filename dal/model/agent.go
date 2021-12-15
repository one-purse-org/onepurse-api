package model

type Agent struct {
	ID string `bson:"id" json:"id"`
}

type AgentAccount struct {
	ID    string `bson:"_id" json:"id"`
	Agent *Agent `bson:"agent" json:"agent"`
	Name  string `bson:"name" json:"name"`
}
