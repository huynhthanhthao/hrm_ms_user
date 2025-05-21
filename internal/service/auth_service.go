package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/huynhthanhthao/hrm_user_service/ent"
	"github.com/huynhthanhthao/hrm_user_service/ent/account"
	"github.com/huynhthanhthao/hrm_user_service/ent/user"
	"github.com/huynhthanhthao/hrm_user_service/internal/dto"
	"github.com/huynhthanhthao/hrm_user_service/internal/helper"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	hrPb "github.com/longgggwwww/hrm-ms-hr/ent/proto/entpb"
)

type AuthService struct {
	client     *ent.Client
	hrClients  *HRServiceClients
	perClients *PermissionServiceClients
}

func NewAuthService(
	client *ent.Client,
	hrClients *HRServiceClients,
	perClients *PermissionServiceClients,
) (*AuthService, error) {
	return &AuthService{
		client:     client,
		hrClients:  hrClients,
		perClients: perClients,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, c *gin.Context, input dto.RegisterInput) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Bắt đầu transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Tạo người dùng mới
	usr, err := tx.User.
		Create().
		SetFirstName(input.FirstName).
		SetLastName(input.LastName).
		SetEmail(input.Email).
		SetAvatar(input.Avatar).
		SetPhone(input.Phone).
		SetWardCode(input.WardCode).
		SetAddress(input.Address).
		SetGender(user.Gender(input.Gender)).
		Save(ctx)

	if err != nil {
		_ = tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
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
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	if err := tx.Commit(); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr.Edges.Account = acc

	c.JSON(http.StatusOK, gin.H{
		"user": usr,
	})
}

func (s *AuthService) Login(ctx context.Context, c *gin.Context, input dto.LoginInput) {
	// Tìm tài khoản theo tên đăng nhập
	acc, err := s.client.Account.
		Query().
		Where(account.UsernameEQ(input.Username)).
		Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// So sánh mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(input.Password)); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Lấy user từ account
	usr, err := acc.QueryUser().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Gọi gRPC lấy employee theo user ID (int -> string)
	employee, err := s.hrClients.HrExt.GetEmployeeByUserId(ctx, &hrPb.GetEmployeeByUserIdRequest{
		UserId: strconv.Itoa(usr.ID),
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Parse duration
	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	// Tạo token
	accessToken, err := GenerateToken(usr.ID, employee.Id, employee.OrgId, accessDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	refreshToken, err := GenerateToken(usr.ID, employee.Id, employee.OrgId, refreshDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	jsonEmployee, err := protojson.Marshal(employee)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("failed to marshal employee: %v", err))
		return
	}

	var employeeMap map[string]interface{}
	if err := json.Unmarshal(jsonEmployee, &employeeMap); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("failed to unmarshal employee JSON: %v", err))
		return
	}

	usr.Edges.Account = acc

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          usr,
		"employee":      employeeMap,
	})
}

func (s *AuthService) DecodeToken(ctx context.Context, token string, c *gin.Context) {
	// Parse the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("#1 DecodeToken: unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("#2 DecodeToken: invalid token claims"))
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("#3 DecodeToken: user_id not found in token claims"))
		return
	}

	userID := int(userIDFloat)

	// Query the user by ID
	usr, err := s.client.User.Query().Where(user.IDEQ(userID)).Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	// Query the account by user
	acc, err := usr.QueryAccount().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr.Edges.Account = acc

	employee, err := s.hrClients.HrExt.GetEmployeeByUserId(ctx, &hrPb.GetEmployeeByUserIdRequest{
		UserId: strconv.Itoa(usr.ID),
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	jsonEmployee, err := protojson.Marshal(employee)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("#4 DecodeToken: failed to marshal employee: %v", err))
		return
	}

	var employeeMap map[string]interface{}
	if err := json.Unmarshal(jsonEmployee, &employeeMap); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest,
			fmt.Errorf("#5 DecodeToken: failed to unmarshal employee JSON: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     usr,
		"employee": employeeMap,
	})
}

func GenerateToken(userID int, employeeID int64, orgID int64, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID,
		"org_id":      orgID,
		"employee_id": employeeID,
		"exp":         time.Now().Add(duration).Unix(),
		"iss":         os.Getenv("ISS_KEY"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}
	return token.SignedString([]byte(secret))
}
