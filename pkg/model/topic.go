package model

type Topic struct {
	Name string `json:"name"`
}

func CreateTopic(name string) *Topic {
	return &Topic{
		Name: name,
	}
}
