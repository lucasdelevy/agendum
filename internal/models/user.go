package models

type User struct {
	Username  string   `json:"username"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	UserType  string   `json:"userType"`
	TeamIDs   []string `json:"teamIds"`
}