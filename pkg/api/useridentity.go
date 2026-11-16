package api

// UserIdentity is who a user is, as other members of an organisation see
// them: enough to show and find them, never credentials, tokens or profile
// settings. Members, team members and other listings of people share it.
// Its JSON names match models.User (firstname, lastname).
type UserIdentity struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
}
