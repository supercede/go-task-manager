package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-app/util"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type tokenService struct{}

func NewToken() *tokenService {
	return &tokenService{}
}

type TokenInterface interface {
	CreateToken(userId uint) (*TokenDetails, error)
	GetTokenMetadata(*http.Request) (*AccessDetails, error)
	// ExtractTokenMetadata(*http.Request) (*AccessDetails, error)
}

func (t *tokenService) CreateToken(userId uint) (*TokenDetails, error) {
	conf, err := util.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to Read config file")
	}
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 60).Unix() //expires after 30 min
	td.TokenUuid = uuid.NewV4().String()

	atSecret := conf.AccessSecret
	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.TokenUuid
	// atClaims["user_id"] = strconv.FormatUint(uint64(userId), 10)
	atClaims["user_id"] = userId
	atClaims["exp"] = td.AtExpires

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	td.AccessToken, err = at.SignedString([]byte(atSecret))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create access token")
	}

	//Creating Refresh Token
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + strconv.FormatUint(uint64(userId), 10)

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	// rtClaims["user_id"] = strconv.FormatUint(uint64(userId), 10)
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	td.RefreshToken, err = rt.SignedString([]byte(conf.RefreshSecret))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	tokenArr := strings.Split(bearerToken, " ")

	if len(tokenArr) == 2 {
		return tokenArr[1]
	}

	return ""
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	conf, err := util.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to Read config file")
	}

	tokenStr := extractToken(r)
	if tokenStr == "" {
		return nil, errors.New("Token not found")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(conf.AccessSecret), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error verifying token")
	}
	return token, nil
}

func CheckTokenValidity(r *http.Request) error {
	token, err := verifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); ok && token.Valid {
		return nil
	}
	return nil
}

func (t *tokenService) GetTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := verifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUUID, ok := claims["access_uuid"].(string)
		userId, userOk := claims["user_id"].(float64)
		uintUserId := uint(userId)
		if ok == false || userOk == false {
			return nil, errors.New("unauthorized")
		}

		return &AccessDetails{
			TokenUuid: accessUUID,
			UserId:    uintUserId,
		}, nil
	}
	return nil, errors.New("Error extracting token data")
}
