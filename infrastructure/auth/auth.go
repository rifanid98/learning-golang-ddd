package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type AuthInterface interface {
	CreateAuth(uint64, *TokenDetails) error
	FetchAuth(string) (uint64, error)
	DeleteRefresh(string) error
	DeleteTokens(*AccessDetails) error
}

type ClientData struct {
	client *redis.Client
}

type AccessDetails struct {
	TokenUuid string
	UserId    uint64
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	TokenUuid    string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

var _ AuthInterface = &ClientData{}

func NewAuthHandler(client *redis.Client) *ClientData {
	return &ClientData{client: client}
}

func (cd *ClientData) CreateAuth(uId uint64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	atCreated, err := cd.client.Set(cd.client.Context(), td.TokenUuid, strconv.Itoa(int(uId)), at.Sub(now)).Result()
	if err != nil {
		return err
	}

	rtCreated, err := cd.client.Set(cd.client.Context(), td.RefreshUuid, strconv.Itoa(int(uId)), rt.Sub(now)).Result()
	if err != nil {
		return err
	}

	if atCreated == "0" || rtCreated == "0" {
		return errors.New("no record inserted")
	}

	return nil
}

func (cd *ClientData) FetchAuth(tokenUuid string) (uint64, error) {
	uId, err := cd.client.Get(cd.client.Context(), tokenUuid).Result()
	if err != nil {
		return 0, err
	}
	uID, err := strconv.ParseUint(uId, 10, 64)
	if err != nil {
		return 0, nil
	}
	return uID, nil
}

func (cd *ClientData) DeleteRefresh(refUuid string) error {
	// delete refresh token
	deleted, err := cd.client.Del(cd.client.Context(), refUuid).Result()
	if err != nil || deleted != 1 {
		return err
	}
	return nil
}

func (cd *ClientData) DeleteTokens(ad *AccessDetails) error {
	// get the refresh uuid
	refUuid := fmt.Sprintf("%s++%d", ad.TokenUuid, ad.UserId)
	// delete access token
	deletedAt, err := cd.client.Del(cd.client.Context(), ad.TokenUuid).Result()
	if err != nil {
		return err
	}
	// delete refresh token
	deletedRt, err := cd.client.Del(cd.client.Context(), refUuid).Result()
	if err != nil {
		return err
	}
	// when the record is deleted, the return value is 1
	if deletedAt != 1 || deletedRt != 1 {
		return errors.New("someting went wrong")
	}
	return nil
}
