package user

type User struct {
	Name string `bson:"_id"`
	Keys []string
}
