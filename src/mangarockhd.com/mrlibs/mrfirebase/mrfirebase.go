package mrfirebase

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	jsoniter "github.com/json-iterator/go"

	firebase "firebase.google.com/go"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"mangarockhd.com/mrlibs/mrconstants"
	"mangarockhd.com/mrlibs/mrutils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var firebaseApp *firebase.App
var backgroundCtx = context.Background()

var firebaseHTTPClient = http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   100 * time.Second,
			KeepAlive: 100 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	Timeout: time.Duration(10 * time.Second),
}

type requestTokenPayload struct {
	// A Firebase Auth custom token from which to create an ID and refresh token pair.
	CustomToken string `json:"token"`
	// Whether or not to return an ID and refresh token. Should always be true.
	ReturnSecureToken bool `json:"returnSecureToken"`
}

type responseTokenPayload struct {
	// Kind         string `json:"kind"`
	IDToken string `json:"idtoken"`
	// RefreshToken string `json:"refreshToken"`
	// ExpiresIn    string `json:"expiresIn"`
}

type tokenCache struct {
	userID   string
	expireAt uint32
}

var cacheTokenUID = make(map[uint64]*tokenCache)
var cacheTokenUIDLock sync.RWMutex

func initCheckCacheToken() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			now := uint32(time.Now().Unix())
			expiredKeys := []uint64{}
			cacheTokenUIDLock.RLock()
			if len(cacheTokenUID) > 0 {
				for key, val := range cacheTokenUID {
					if val.expireAt < now {
						expiredKeys = append(expiredKeys, key)
					}
				}
			}
			cacheTokenUIDLock.RUnlock()
			total := len(expiredKeys)
			if total > 0 {
				cacheTokenUIDLock.Lock()
				for i := 0; i < total; i++ {
					delete(cacheTokenUID, expiredKeys[i])
				}
				cacheTokenUIDLock.Unlock()
			}
		}
	}()
}

// Initialize init module
func Initialize(credential string) {
	creds, err := google.CredentialsFromJSON(backgroundCtx, []byte(credential),
		mrconstants.FIRESTORE_SCOPE, mrconstants.FIREBASE_MESSAGE_SCOPE,
		mrconstants.FIREBASE_USER_INFO_EMAIL_SCOPE,
		mrconstants.FIREBASE_IDENTITY_TOOLKIT_SCOPE,
	)
	if err != nil {
		log.Panic("error reading Firebase credential: ", err)
	}

	firebaseOpt := option.WithCredentials(creds)
	firebaseApp, err = firebase.NewApp(backgroundCtx, nil, firebaseOpt)
	if err != nil {
		log.Panic("error initializing app: ", err)
	}
	go initCheckCacheToken()
}

// GetUserIDFromToken get uid from idToken or cookieToken
func GetUserIDFromToken(token string) string {

	hash := xxhash.Checksum64([]byte(token))
	cacheTokenUIDLock.RLock()
	cached := cacheTokenUID[hash]
	cacheTokenUIDLock.RUnlock()
	if cached != nil {
		return cached.userID
	}

	dotIndex := strings.Index(token, ".")
	if dotIndex == -1 {
		return ""
	} else {
		authClient, err := firebaseApp.Auth(backgroundCtx)
		if err != nil {
			mrutils.Log("GetUserIdFromToken error 1 %v", err)
			return ""
		}
		if dotIndex > 80 { // id token
			tokenDecoded, err := authClient.VerifyIDToken(backgroundCtx, token)
			if tokenDecoded == nil || err != nil {
				mrutils.Log("GetUserIdFromToken error 2 %v", err)
				return ""
			}
			cacheItem := &tokenCache{
				expireAt: uint32(tokenDecoded.Expires),
				userID:   tokenDecoded.UID,
			}
			cacheTokenUIDLock.Lock()
			cacheTokenUID[hash] = cacheItem
			cacheTokenUIDLock.Unlock()
			return tokenDecoded.UID
		}
	}
	return ""
}

// GetUserIDFromRequest get firebase user id from request header
func GetUserIDFromRequest(r *http.Request) string {
	idToken := r.Header.Get("Authorization")
	if len(idToken) <= 40 {
		return ""
	}

	return GetUserIDFromToken(idToken)
}

func getCustomToken(userID string) (string, error) {
	if len(userID) == 0 {
		mrutils.Log("getCustomToken - Invalid userID")
		return "", errors.New("Invalid userID")
	}

	authClient, err := firebaseApp.Auth(backgroundCtx)
	if err != nil {
		mrutils.Log("getCustomIDToken error create auth's client %v", err)
		return "", err
	}

	customToken, err := authClient.CustomToken(backgroundCtx, userID)
	if err != nil {
		mrutils.Log("create custom token error=%v", err)
		return "", err
	}

	return customToken, nil
}

// GetIDTokenFromUserID get idToken from userID
func GetIDTokenFromUserID(userID, firebaseAPIKey string) (string, error) {
	if len(userID) == 0 || len(firebaseAPIKey) == 0 {
		mrutils.Log("GetIDTokenFromUserID invalid userID or firebase API key")
		return "", errors.New("Invalid useriD or firebase API key")
	}

	customToken, err := getCustomToken(userID)
	if err != nil {
		return "", err
	}

	reqPayload := requestTokenPayload{}
	reqPayload.CustomToken = customToken
	reqPayload.ReturnSecureToken = true
	payload, err := json.Marshal(reqPayload)
	if err != nil {
		mrutils.Log("GetIDTokenFromUserID error encoded request payload %v", err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, mrconstants.FIREBASE_VERIFY_CUSTOM_TOKEN_URL+firebaseAPIKey, bytes.NewBuffer(payload))
	if err != nil {
		mrutils.Log("GetIDTokenFromUserID create new request error=%v", err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := firebaseHTTPClient.Do(req)
	if err != nil {
		mrutils.Log("GetIDTokenFromUserID do request error=%v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mrutils.Log("GetIDTokenFromUserID read response body err=%v", err)
		return "", err
	}

	respTokenPayload := responseTokenPayload{}
	err = json.Unmarshal(body, &respTokenPayload)
	if err != nil {
		mrutils.Log("GetIDTokenFromUserID decoded response payload err=%v", err)
		return "", err
	}

	return respTokenPayload.IDToken, nil
}

// GetUserEmailFromUserID get user's email by uid
func GetUserEmailFromUserID(userID string) (string, error) {
	if len(userID) == 0 {
		return "", errors.New("Invalid userID")
	}

	authClient, err := firebaseApp.Auth(backgroundCtx)
	if err != nil {
		mrutils.Log("getCustomIDToken error create auth's client %v", err)
		return "", err
	}

	userRecord, err := authClient.GetUser(backgroundCtx, userID)
	if err != nil {
		return "", err
	}

	if userRecord == nil {
		return "", nil
	}

	return userRecord.Email, nil
}
