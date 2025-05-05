package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
	"user/ent"
	"user/ent/account"
	"user/ent/user"
	"user/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	client *ent.Client
}

func NewUserService(client *ent.Client) *UserService {
	return &UserService{
		client: client,
	}
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
	CompanyId string
}

func (s *UserService) Register(ctx context.Context, c *gin.Context, input RegisterInput) (*dto.UserResponse, error) {
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
		SetCompanyID(input.CompanyId).
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

	return &dto.UserResponse{
			ID:        usr.ID,
			FirstName: usr.FirstName,
			LastName:  usr.LastName,
			Email:     usr.Email,
			Phone:     usr.Phone,
			Address:   usr.Address,
			WardCode:  usr.WardCode,
			CompanyId: usr.CompanyID,
			Gender:    string(usr.Gender),
			Account: &dto.AccountResponse{
				Username: acc.Username,
			},
		},
		nil
}

type LoginInput struct {
	Username string
	Password string
}

func (s *UserService) Login(ctx context.Context, c *gin.Context, input LoginInput) (*dto.LoginResponse, error) {
	// Tìm tài khoản theo tên đăng nhập
	acc, err := s.client.Account.
		Query().
		Where(account.UsernameEQ(input.Username)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Tài khoản không tồn tại: %v", http.StatusNotFound, err)
	}

	// Kiểm tra mật khẩu
	err = bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password))
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Mật khẩu không đúng: %v", http.StatusUnauthorized, err)
	}

	// Lấy thông tin người dùng
	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Không thể lấy thông tin người dùng: %v", http.StatusInternalServerError, err)
	}

	// Tạo token
	accessTokenDuration, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshTokenDuration, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	accessToken, err := generateToken(acc.ID, accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo access token: %v", http.StatusInternalServerError, err)
	}

	refreshToken, err := generateToken(acc.ID, refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d: Lỗi tạo refresh token: %v", http.StatusInternalServerError, err)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			ID:        usr.ID,
			FirstName: usr.FirstName,
			LastName:  usr.LastName,
			Email:     usr.Email,
			Phone:     usr.Phone,
			Address:   usr.Address,
			CompanyId: usr.CompanyID,
			WardCode:  usr.WardCode,
			Gender:    string(usr.Gender),
			Account: &dto.AccountResponse{
				Username: acc.Username,
			},
		},
	}, nil
}

func generateToken(accountID int, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"account_id": accountID,
		"exp":        time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
