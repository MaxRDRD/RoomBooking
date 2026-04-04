package auth

import (
	myerrors "RoomBookingService/pkg/my_errors"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AdminDummyUserID = "11111111-1111-1111-1111-111111111111"
	UserDummyUserID  = "22222222-2222-2222-2222-222222222222"
)

type TokenService interface {
	GenerateDummyToken(role string) (string, error)
	GenerateToken(userID uuid.UUID, role string) (string, error)
	ParseToken(tokenString string) (uuid.UUID, string, error)
}

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	return &JWTService{
		secret: []byte(secret),
		ttl:    24 * time.Hour,
	}
}

func (s *JWTService) GenerateDummyToken(role string) (string, error) {
	role = strings.ToLower(strings.TrimSpace(role))

	var userID string
	switch role {
	case "admin":
		userID = AdminDummyUserID
	case "user":
		userID = UserDummyUserID
	default:
		return "", myerrors.ErrInvalidRole
	}

	return s.signToken(userID, role)
}


func (s *JWTService) GenerateToken(userID uuid.UUID, role string) (string, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "admin" && role != "user" {
		return "", myerrors.ErrInvalidRole
	}
	if userID == uuid.Nil {
		return "", myerrors.ErrInvalidCredentials
	}

	return s.signToken(userID.String(), role)
}

func (s *JWTService) signToken(userID string, role string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseToken(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, "", myerrors.ErrInvalidCredentials
	}

	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, "", err
	}

	return parsedUserID, claims.Role, nil
}
