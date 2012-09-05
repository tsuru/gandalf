package repository

type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}
