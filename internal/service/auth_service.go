package service

import (
	"context"
	"errors"
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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	hrPb "github.com/longgggwwww/hrm-ms-hr/ent/proto/entpb"
	permPb "github.com/longgggwwww/hrm-ms-permission/ent/proto/entpb"
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

// Struct kết quả trả về cho login
type LoginResult struct {
	User       *ent.User
	Account    *ent.Account
	Employee   map[string]interface{}
	Roles      []map[string]interface{}
	Perms      []map[string]interface{}
	PermCodes  []string
	EmployeeID *int64
	OrgID      *int64
}

// Lấy account theo username
func (s *AuthService) getAccountByUsername(ctx context.Context, username string) (*ent.Account, error) {
	return s.client.Account.
		Query().
		Where(account.UsernameEQ(username)).
		Only(ctx)
}

// Kiểm tra password
func checkPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Lấy user từ account
func getUserFromAccount(ctx context.Context, acc *ent.Account) (*ent.User, error) {
	return acc.QueryUser().Only(ctx)
}

// Lấy roles, perms, permCodes
func (s *AuthService) getUserRolesPerms(ctx context.Context, userID int) ([]map[string]interface{}, []map[string]interface{}, []string, error) {
	userIDStr := strconv.Itoa(userID)
	permsResp, err := s.perClients.PermExt.GetUserPerms(ctx, &permPb.GetUserPermsRequest{UserId: userIDStr})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get user perms: %w", err)
	}
	rolesResp, err := s.perClients.PermExt.GetUserRoles(ctx, &permPb.GetUserRolesRequest{UserId: userIDStr})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	permCodeSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		for _, p := range r.Perms {
			permCodeSet[p.Code] = struct{}{}
		}
	}
	permCodes := make([]string, 0, len(permCodeSet))
	for code := range permCodeSet {
		permCodes = append(permCodes, code)
	}
	return helper.ToRoleArr(rolesResp.Roles), helper.ToPermArr(permsResp.Perms), permCodes, nil
}

// Lấy employee, employeeID, orgID
func (s *AuthService) getEmployeeInfo(ctx context.Context, userID int) (*hrPb.Employee, *int64, *int64) {
	var employeeID, orgID *int64
	employee, err := s.hrClients.HrExt.GetEmployeeByUserId(ctx, &hrPb.GetEmployeeByUserIdRequest{
		UserId: strconv.Itoa(userID),
	})

	if err == nil && employee != nil {
		employeeID = &employee.Id
		orgID = &employee.OrgId
	}
	return employee, employeeID, orgID
}

func (s *AuthService) Login(ctx context.Context, c *gin.Context, input dto.LoginInput) {
	acc, err := s.getAccountByUsername(ctx, input.Username)

	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	if acc.Status == account.StatusInactive {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("#1 Login: account is inactive"))
		return
	}

	if err := checkPassword(acc.Password, input.Password); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	usr, err := getUserFromAccount(ctx, acc)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}
	rolesArr, permsArr, permCodes, err := s.getUserRolesPerms(ctx, usr.ID)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}
	employee, employeeID, orgID := s.getEmployeeInfo(ctx, usr.ID)

	var employeeMap map[string]interface{}
	if employee != nil {
		employeeMap = helper.ToEmployeeMap(employee)
	}

	usr.Edges.Account = acc

	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	accessToken, err := GenerateAccessToken(TokenClaimsInput{
		UserID:         usr.ID,
		EmployeeID:     employeeID,
		EmployeeStatus: employeeMap["status"].(string),
		OrgID:          orgID,
		Duration:       accessDur,
		Perms:          permCodes,
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	refreshToken, err := GenerateRefreshToken(usr.ID, refreshDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          usr,
		"employee":      employeeMap,
		"roles":         rolesArr,
		"perms":         permsArr,
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

	// Validate account status
	if acc.Status == account.StatusInactive {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("#4 DecodeToken: account is inactive"))
		return
	}
	usr.Edges.Account = acc

	// Query employee
	var employeeMap map[string]interface{}
	employee, _, _ := s.getEmployeeInfo(ctx, usr.ID)
	if employee != nil {
		employeeMap = helper.ToEmployeeMap(employee)
	}

	// Query roles & perms
	userIDStr := strconv.Itoa(usr.ID)
	permsResp, err := s.perClients.PermExt.GetUserPerms(ctx, &permPb.GetUserPermsRequest{UserId: userIDStr})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("#4 DecodeToken: failed to get user perms: %w", err))
		return
	}

	rolesResp, err := s.perClients.PermExt.GetUserRoles(ctx, &permPb.GetUserRolesRequest{UserId: userIDStr})
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, fmt.Errorf("#5 DecodeToken: failed to get user roles: %w", err))
		return
	}

	// Lấy tất cả perm_codes từ các role (gộp, loại trùng)
	permCodeSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		for _, p := range r.Perms {
			permCodeSet[p.Code] = struct{}{}
		}
	}
	permCodes := make([]string, 0, len(permCodeSet))
	for code := range permCodeSet {
		permCodes = append(permCodes, code)
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     usr,
		"employee": employeeMap,
		"roles":    helper.ToRoleArr(rolesResp.Roles),
		"perms":    helper.ToPermArr(permsResp.Perms),
	})
}

// Định nghĩa struct chứa thông tin để sign token
type TokenClaimsInput struct {
	UserID         int
	EmployeeID     *int64
	EmployeeStatus string
	OrgID          *int64
	Duration       time.Duration
	Roles          []string
	Perms          []string
}

func GenerateAccessToken(input TokenClaimsInput) (string, error) {
	claims := jwt.MapClaims{
		"user_id":         input.UserID,
		"org_id":          input.OrgID,
		"employee_status": input.EmployeeStatus,
		"employee_id":     input.EmployeeID,
		"exp":             time.Now().Add(input.Duration).Unix(),
		"iss":             os.Getenv("ISS_KEY"),
		"perm_codes":      input.Perms,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}
	return token.SignedString([]byte(secret))
}

// Generate refresh token: chỉ chứa user_id và exp
func GenerateRefreshToken(userID int, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iss":     os.Getenv("ISS_KEY"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}
	return token.SignedString([]byte(secret))
}

// POST /auth/refresh-token
func (s *AuthService) RefreshToken(ctx context.Context, c *gin.Context, refreshToken string) {
	if refreshToken == "" {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("refresh token missing"))
		return
	}

	parsedToken, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !parsedToken.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("refresh token expired"))
			return
		}
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("invalid refresh token: %v", err))
		return
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("invalid token claims"))
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("user_id not found in token claims"))
		return
	}
	userID := int(userIDFloat)
	usr, err := s.client.User.Query().Where(user.IDEQ(userID)).Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	// Query account to check status
	acc, err := usr.QueryAccount().Only(ctx)
	if err != nil {
		helper.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	if acc.Status == account.StatusInactive {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("account is inactive"))
		return
	}

	// Get employee info for claims
	employee, employeeID, orgID := s.getEmployeeInfo(ctx, usr.ID)
	var employeeStatus string
	if employee != nil {
		employeeMap := helper.ToEmployeeMap(employee)
		employeeStatus = employeeMap["status"].(string)
	}

	// Query roles & perms
	userIDStr := strconv.Itoa(usr.ID)
	rolesResp, err := s.perClients.PermExt.GetUserRoles(ctx, &permPb.GetUserRolesRequest{UserId: userIDStr})
	if err != nil {
		helper.RespondWithError(c, http.StatusUnauthorized, fmt.Errorf("failed to get user roles: %w", err))
		return
	}

	// Lấy tất cả perm_codes từ các role (gộp, loại trùng)
	permCodeSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		for _, p := range r.Perms {
			permCodeSet[p.Code] = struct{}{}
		}
	}
	permCodes := make([]string, 0, len(permCodeSet))
	for code := range permCodeSet {
		permCodes = append(permCodes, code)
	}

	accessDur, _ := time.ParseDuration(os.Getenv("JWT_ACCESS_TOKEN_DURATION"))
	refreshDur, _ := time.ParseDuration(os.Getenv("JWT_REFRESH_TOKEN_DURATION"))

	// Generate new access token
	accessToken, err := GenerateAccessToken(TokenClaimsInput{
		UserID:         usr.ID,
		EmployeeID:     employeeID,
		EmployeeStatus: employeeStatus,
		OrgID:          orgID,
		Duration:       accessDur,
		Perms:          permCodes,
	})
	if err != nil {
		helper.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	// Generate new refresh token (recommended for security)
	newRefreshToken, err := GenerateRefreshToken(usr.ID, refreshDur)
	if err != nil {
		helper.RespondWithError(c, http.StatusUnauthorized, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(accessDur.Seconds()),
	})
}
