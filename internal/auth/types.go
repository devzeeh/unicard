package authentication

// Create a struct to catch the incoming JSON from the frontend
type SignupRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	CardNumber    string `json:"cardNumber"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ContactNumber string `json:"contactNumber"`
}

// Create a standard API response struct
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// User struct to hold signup data (Keep your existing one)
type User struct {
	UserID     string
	Username   string
	Fullname   string
	Email      string
	Phone      string
	CardNumber string
	Password   string
	CardID     string
	Usertype   string
	Balance    float64
	CreatedAt  string
}