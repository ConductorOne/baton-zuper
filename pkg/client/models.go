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

// Create Users.
type WorkHour struct {
	Day           string `json:"day"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	WorkMins      int    `json:"work_mins"`
	TrackLocation bool   `json:"track_location"`
	IsEnabled     string `json:"is_enabled"`
}

type UserPayload struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Designation string `json:"designation"`
	EmpCode     string `json:"emp_code"`
	RoleID      string `json:"role_id"`
}

type CreateUserRequest struct {
	WorkHours []WorkHour  `json:"work_hours"`
	User      UserPayload `json:"user"`
}

type CreateUserResponse struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Data    struct {
		UserUID string `json:"user_uid"`
	} `json:"data"`
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

// Teams Models.
type CreatedBy struct {
	UserUID           string `json:"user_uid"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	ExternalLoginID   string `json:"external_login_id"`
	HomePhoneNumber   string `json:"home_phone_number"`
	Designation       string `json:"designation"`
	EmpCode           string `json:"emp_code"`
	Prefix            string `json:"prefix"`
	WorkPhoneNumber   string `json:"work_phone_number"`
	MobilePhoneNumber string `json:"mobile_phone_number"`
	ProfilePicture    string `json:"profile_picture"`
	HourlyLaborCharge string `json:"hourly_labor_charge"`
	IsActive          bool   `json:"is_active"`
	IsDeleted         bool   `json:"is_deleted"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type Team struct {
	TeamUID         string    `json:"team_uid"`
	TeamName        string    `json:"team_name"`
	TeamColor       string    `json:"team_color"`
	TeamDescription string    `json:"team_description"`
	TeamTimezone    string    `json:"team_timezone"`
	UserCount       int       `json:"user_count"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
	CreatedBy       CreatedBy `json:"created_by"`
}

type TeamsResponse struct {
	Type         string `json:"type"`
	Data         []Team `json:"data"`
	TotalRecords int    `json:"total_records"`
	CurrentPage  int    `json:"current_page"`
	TotalPages   int    `json:"total_pages"`
}

type TeamDetailsResponse struct {
	Type string `json:"type"`
	Data Team   `json:"data"`
}

type TeamDetailsWithUsersResponse struct {
	Type string `json:"type"`
	Data struct {
		Team  Team        `json:"team"`
		Users []ZuperUser `json:"users"`
	} `json:"data"`
}
