package gandalf

type user struct {
	Name string `bson:"_id"`
	Key  []string
}

type repository struct {
	Name string
	User []string
}
