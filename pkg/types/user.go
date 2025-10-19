package types

import "fmt"

// User represents a simulated user of the system
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TeamName string `json:"team_name"`
}

// Role constants
const (
	RoleComplianceManager = "compliance_manager"
	RoleEngineer          = "engineer"
)

// Predefined users
var (
	UserAlice = &User{
		ID:       "user-001",
		Name:     "Alice Johnson",
		Email:    "alice@example.com",
		Role:     RoleComplianceManager,
		TeamName: "Compliance",
	}
	UserBob = &User{
		ID:       "user-002",
		Name:     "Bob Martinez",
		Email:    "bob@example.com",
		Role:     RoleEngineer,
		TeamName: "Engineering",
	}
	UserCarol = &User{
		ID:       "user-003",
		Name:     "Carol Zhang",
		Email:    "carol@example.com",
		Role:     RoleEngineer,
		TeamName: "Engineering",
	}
)

// AllUsers returns all predefined users
func AllUsers() []*User {
	return []*User{UserAlice, UserBob, UserCarol}
}

// ValidateUser checks if a User meets all validation rules
func ValidateUser(u *User) error {
	if u == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Validate ID
	if u.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Validate name
	if u.Name == "" {
		return fmt.Errorf("user name cannot be empty")
	}

	// Validate role
	validRoles := []string{RoleComplianceManager, RoleEngineer}
	valid := false
	for _, r := range validRoles {
		if u.Role == r {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid role: %s, must be one of %v", u.Role, validRoles)
	}

	return nil
}

// CanViewTechnicalDetails returns true if the user role can view technical details
func (u *User) CanViewTechnicalDetails() bool {
	return u.Role == RoleEngineer
}

// GetUserByRole returns the first user with the specified role
func GetUserByRole(role string) *User {
	for _, u := range AllUsers() {
		if u.Role == role {
			return u
		}
	}
	return nil
}
