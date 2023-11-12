package dataStore

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id			primitive.ObjectID 	`bson:"_id"`
	Username	string 				`bson:"username"`
	Password 	string 				`bson:"password"`
	Admin		bool 				`bson:"admin"`
	Tokens 		[]string 			`bson:"tokens"`
	Activated	bool				`bson:"activated"`
}