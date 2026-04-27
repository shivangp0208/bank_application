package validator

import (
	"fmt"
	"net/mail"
	"regexp"

	"github.com/shivangp0208/bank_application/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z]+(\s[a-zA-Z]+)+$`).MatchString
)

func ValidateString(str string, minLen int, maxLen int) error {
	n := len(str)
	if n < minLen || n > maxLen {
		return fmt.Errorf("length must be between %d-%d", minLen, maxLen)
	}
	return nil
}

func ValidateUsername(username string) error {
	if err := ValidateString(username, 2, 100); err != nil {
		return err
	}
	if !isValidUsername(username) {
		return fmt.Errorf("must contain only lowercase alphabets, underscore and digits")
	}
	return nil
}

func ValidateSecretCode(secretCode string) error {
	if err := ValidateString(secretCode, 36, 36); err != nil {
		return err
	}
	return nil
}

func ValidateFullName(name string) error {
	if err := ValidateString(name, 1, 255); err != nil {
		return err
	}
	if !isValidFullName(name) {
		return fmt.Errorf("must contain only alphabets, spaces")
	}
	return nil
}

func ValidatePassword(password string) error {
	return ValidateString(password, 8, 255)
}

func ValidateEmail(email string) error {
	if err := ValidateString(email, 10, 255); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

func ValidateCreateUserReq(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := ValidateUsername(req.Username); err != nil {
		violations = append(violations, ValidateField(req.Username, err))
	}
	if err := ValidateEmail(req.Email); err != nil {
		violations = append(violations, ValidateField(req.Email, err))
	}
	if err := ValidateFullName(req.FullName); err != nil {
		violations = append(violations, ValidateField(req.FullName, err))
	}
	if err := ValidatePassword(req.Password); err != nil {
		violations = append(violations, ValidateField(req.Password, err))
	}
	return violations
}

func ValidateUpdateUserReq(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := ValidateUsername(req.Username); err != nil {
		violations = append(violations, ValidateField(req.Username, err))
	}
	if req.Email != nil && len(*req.Email) > 0 {
		if err := ValidateEmail(*req.Email); err != nil {
			violations = append(violations, ValidateField(*req.Email, err))
		}
	}
	if req.FullName != nil && len(*req.FullName) > 0 {
		if err := ValidateFullName(*req.FullName); err != nil {
			violations = append(violations, ValidateField(*req.FullName, err))
		}
	}
	if req.Password != nil && len(*req.Password) > 0 {
		if err := ValidatePassword(*req.Password); err != nil {
			violations = append(violations, ValidateField(*req.Password, err))
		}
	}
	return violations
}

func ValidateLoginUserReq(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := ValidateUsername(req.Username); err != nil {
		violations = append(violations, ValidateField(req.Username, err))
	}
	if err := ValidatePassword(req.Password); err != nil {
		violations = append(violations, ValidateField(req.Password, err))
	}
	return violations
}

func ValidateVerifyUserEmailReq(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := ValidateUsername(req.Username); err != nil {
		violations = append(violations, ValidateField(req.Username, err))
	}
	if err := ValidateSecretCode(req.SecretCode); err != nil {
		violations = append(violations, ValidateField(req.SecretCode, err))
	}
	return violations
}
