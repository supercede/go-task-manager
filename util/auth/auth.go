package auth

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type AuthInterface interface {
	CreateAuth(uint, *TokenDetails) error
	FetchAuth(string) (string, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
}

type service struct {
	client *redis.Client
}

// var _ AuthInterface = &service{}

func NewAuth(client *redis.Client) *service {
	return &service{client: client}
}

type AccessDetails struct {
	TokenUuid string
	UserId    uint
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

//Save token metadata to Redis
func (tk *service) CreateAuth(userId uint, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	atCreated, err := tk.client.Set(td.TokenUuid, userId, at.Sub(now)).Result()
	if err != nil {
		return err
	}
	rtCreated, err := tk.client.Set(td.RefreshUuid, userId, rt.Sub(now)).Result()
	if err != nil {
		return err
	}
	if atCreated == "0" || rtCreated == "0" {
		return errors.New("no record inserted")
	}
	return nil
}

//Check the metadata saved
func (tk *service) FetchAuth(tokenUuid string) (string, error) {
	userid, err := tk.client.Get(tokenUuid).Result()
	if err != nil {
		return "", err
	}
	return userid, nil
}

//Once a user row in the token table
func (tk *service) DeleteTokens(authD *AccessDetails) error {
	//get the refresh uuid
	refreshUuid := fmt.Sprintf("%s++%s", authD.TokenUuid, strconv.FormatUint(uint64(authD.UserId), 10))
	//delete access token
	deletedAt, err := tk.client.Del(authD.TokenUuid).Result()
	if err != nil {
		log.Warning("Failed to delete access token")
		return err
	}
	//delete refresh token
	deletedRt, err := tk.client.Del(refreshUuid).Result()
	if err != nil {
		log.Warning("Failed to delete refresh token")
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		log.Warning("Access or Refresh token not deleted")
		return errors.New("something went wrong")
	}
	return nil
}

func (tk *service) DeleteRefresh(refreshUuid string) error {
	//delete refresh token
	deleted, err := tk.client.Del(refreshUuid).Result()
	if err != nil || deleted == 0 {
		return errors.New("Invalid refresh token")
	}
	// delete access token if any
	atUUID := strings.Split(refreshUuid, "++")[0]
	_, err = tk.client.Get(atUUID).Result()
	if err != redis.Nil {
		deletedAt, delErr := tk.client.Del(atUUID).Result()
		if delErr != nil || deletedAt == 0 {
			return errors.New("Failed to delete access token")
		}
	}
	return nil
}
