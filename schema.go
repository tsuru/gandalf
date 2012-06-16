package gandalf

type user struct {
	Name string `bson:"_id"`
	Key  string
}

type project struct {
	Name string
	User string
}
