package service

import (
	"context"
	"fmt"
	"net/http"
	"user/ent"
	"user/ent/account"
	"user/ent/user"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	client *ent.Client
}

type RegisterInput struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Email     string
	Phone     string
	WardCode  string
	Address   string
	Gender    string
}

func NewUserService(client *ent.Client) *UserService {
	return &UserService{
		client: client,
	}
}

func (s *UserService) Register(ctx context.Context, c *gin.Context, input RegisterInput) (*ent.User, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi mã hóa mật khẩu: %v", http.StatusConflict, err)
	}

	// Bắt đầu transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi bắt đầu transaction: %v", http.StatusInternalServerError, err)
	}

	// Tạo người dùng mới
	usr, err := tx.User.
		Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetEmail(input.Email).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetGender(user.Gender(input.Gender)).
		Save(ctx)

	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo người dùng: %v", http.StatusBadRequest, err)
	}

	// Tạo tài khoản cho người dùng
	acc, err := tx.Account. 
		Create().
		SetUsername(input.Username).
		SetPassword(string(hashedPassword)).
		SetStatus(account.StatusActive).
		SetUser(usr).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo tài khoản: %v", http.StatusBadRequest, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi commit giao dịch: %v", http.StatusInternalServerError, err)
	}

	usr.Edges.Account = acc

	return usr, nil
}


