package userclient

type UserCreateDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"   `
	Name     string `json:"name"    `
	Surname  string `json:"surname" `
	Address  string `json:"address" `
	Role     string `json:"role"    `
}

type UserDTO struct {
	Id       uint   `json:"id"      `
	Username string `json:"username"`
	Email    string `json:"email"   `
	Name     string `json:"name"    `
	Surname  string `json:"surname" `
	Address  string `json:"address" `
	Role     string `json:"role"    `
}

type LoginDTO struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
}

type JWTDTO struct {
	Jwt string `json:"jwt"`
}
