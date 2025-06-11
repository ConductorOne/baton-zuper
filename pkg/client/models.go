package client

// Pagination Models.
type pageToken struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type PageOptions struct {
	PageToken string
	PageSize  int
}

// Users Models.
type ZuperUser struct {
	UserUID     string      `json:"user_uid"`
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	Email       string      `json:"email"`
	Designation string      `json:"designation"`
	EmpCode     string      `json:"emp_code"`
	IsActive    bool        `json:"is_active"`
	IsDeleted   bool        `json:"is_deleted"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Role        *Role       `json:"role"`
	AccessRole  *AccessRole `json:"access_role"`
}

type UsersResponse struct {
	Type         string      `json:"type"`
	Data         []ZuperUser `json:"data"`
	TotalRecords int         `json:"total_records"`
	TotalPages   int         `json:"total_pages"`
	CurrentPage  int         `json:"current_page"`
}

type UserDetailsResponse struct {
	Type string    `json:"type"`
	Data ZuperUser `json:"data"`
}

// Error Models.
type ZuperError struct {
	MessageError string `json:"message"`
	Title        string `json:"title"`
	Type         string `json:"type"`
}

// Role & Access Models.
type Role struct {
	RoleUID  string `json:"role_uid"`
	RoleName string `json:"role_name"`
	RoleKey  string `json:"role_key"`
}

type AccessRole struct {
	AccessRoleUID   string `json:"access_role_uid"`
	AccessRoleName  string `json:"role_name"`
	RoleDescription string `json:"role_description"`
}
