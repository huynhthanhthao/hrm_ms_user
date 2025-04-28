package service

import (
	"context"
	"fmt"
	"user/ent"
	"user/ent/account"
	"user/ent/user"

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
	return &UserService{client: client}
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*ent.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
        return nil, fmt.Errorf("Lỗi mã hóa mật khẩu: %v", err)
    }

	tx, err := s.client.Tx(ctx)

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
        if rerr := tx.Rollback(); rerr != nil {
            return nil, fmt.Errorf("Lỗi rollback: %v", rerr)
        }
        return nil, fmt.Errorf("Lỗi tạo người dùng: %v", err)
    }

	acc, err := tx.Account.
        Create().
        SetUsername(input.Username).
        SetPassword(string(hashedPassword)).
        SetStatus(account.StatusActive).
        SetUser(usr). 
        Save(ctx)

	if err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            return nil, fmt.Errorf("Lỗi rollback: %v", rerr)
        }
        return nil, fmt.Errorf("Lỗi tạo tài khoản: %v", err)
    }

	usr, err = tx.User.
        UpdateOne(usr).
        SetAccount(acc).
        Save(ctx)

	if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("Lỗi commit giao dịch: %v", err)
    }	

    return usr, nil	
}