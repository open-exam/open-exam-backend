package main

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/open-exam/open-exam-backend/shared"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"github.com/open-exam/open-exam-backend/util"
	"golang.org/x/crypto/argon2"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	errInvalidHash         = errors.New("the encoded hash is not in the correct format")
	errIncompatibleVersion = errors.New("incompatible version of argon2")
	mode                   = "prod"
	errServiceConnection   = gin.H{
		"error":             "server_error",
		"error_description": "could not connect to internal service",
	}
	errJWTCreation = gin.H{
		"error":             "server_error",
		"error_description": "an error occurred while creating JWTs",
	}
	errBadJWT = gin.H{
		"error":             "access_denied",
		"error_description": "invalid code",
	}
	clientIds          = []string{"api", "exam"}
	defaultRedirectUri string
	userService        string
	jwtPrivateKey      *rsa.PrivateKey
	jwtPublicKey       *rsa.PublicKey
	redisCluster       *redis.ClusterClient
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type tokenSet struct {
	id      string
	access  string
	refresh string
}

func main() {
	shared.SetEnv(&mode)
	gin.SetMode(gin.DebugMode)

	if mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	validateOptions()

	listenAddr := os.Getenv("listen_addr")

	router := gin.New()
	router.Use(gin.Recovery())
	if mode == "dev" {
		router.Use(gin.Logger())
	}

	router.GET("/authorize", authorize)
	router.GET("/token", getToken)
	router.GET("/logout", logout)
	router.GET("/refresh", refresh)

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("failed to start oauth2 server: %v", err)
	}
}

func validateOptions() {
	defaultRedirectUri = os.Getenv("redirect_uri")
	userService = os.Getenv("user_service")

	tempJwtPrivateKey, err := util.DecodeBase64([]byte(os.Getenv("jwt_private_key")))
	if err != nil {
		log.Fatalf("invalid jwt_private_key")
	}
	jwtPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(tempJwtPrivateKey)
	if err != nil {
		log.Fatalf("invalid jwt_public_key")
	}

	tempJwtPublicKey, err := util.DecodeBase64([]byte(os.Getenv("jwt_public_key")))
	if err != nil {
		log.Fatalf("invalid jwt_public_key")
	}
	jwtPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(tempJwtPublicKey)
	if err != nil {
		log.Fatalf("invalid jwt_public_key")
	}

	redisAddrs := util.SplitAndParse(os.Getenv("redis_addrs"))
	redisPassword := os.Getenv("redis_pass")

	redisCluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisAddrs,
		Password: redisPassword,
	})
}

func authorize(ctx *gin.Context) {
	var (
		responseType        = ctx.Query("response_type")
		clientId            = ctx.Query("client_id")
		redirectUri         = ctx.DefaultQuery("redirect_uri", defaultRedirectUri)
		scope               = ctx.Query("scope")
		state               = ctx.Query("state")
		username            = ctx.Query("username")
		passwd              = ctx.Query("password")
		codeChallenge       = ctx.Query("code_challenge")
		codeChallengeMethod = ctx.Query("code_challenge_method")
		responseMethod      = ctx.DefaultQuery("response_method", "redirect")
	)

	if responseType != "code" {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid response_type",
		})
		return
	}

	if codeChallengeMethod != "S256" {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid code_challenge_method",
		})
		return
	}

	if util.IsInList(clientId, &clientIds) == -1 {
		ctx.JSON(400, gin.H{
			"error": "unauthorized_client",
		})
		return
	}

	if len(username) > 0 && len(passwd) > 0 && len(codeChallenge) > 0 {
		conn, err := shared.GetGrpcConn(userService)
		client := pb.NewUserServiceClient(conn)

		response, err := client.FindOne(context.Background(), &pb.FindOneRequest{
			Email:    username,
			Password: true,
		})

		defer conn.Close()

		if err != nil {
			ctx.JSON(500, errServiceConnection)
			return
		}

		if len(response.Password) == 0 {
			ctx.JSON(400, gin.H{
				"error":             "access_denied",
				"error_description": "404; the user does not exist",
			})
			return
		} else {
			if verifyHash(passwd, response.Password) {
				genJwt, err := createJWT(response.Id, 30, scope, "authorize", map[string]string{})
				if err != nil {
					ctx.JSON(500, errJWTCreation)
					return
				}

				res := redisCluster.LPush(ctx, response.Id, codeChallenge, scope)
				redisCluster.Expire(ctx, response.Id, time.Second*30)
				if res.Err() != nil {
					fmt.Println(res.Err())
					ctx.JSON(500, errServiceConnection)
					return
				}

				if responseMethod == "redirect" {
					ctx.Redirect(302, redirectUri+"?code="+url.QueryEscape(genJwt)+"&state="+url.QueryEscape(state))
				} else {
					ctx.JSON(200, gin.H{
						"code":  genJwt,
						"state": state,
					})
					return
				}
			} else {
				ctx.JSON(400, gin.H{
					"error":             "access_denied",
					"error_description": "invalid credentials",
				})
				return
			}
		}
	} else {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid parameters",
		})
		return
	}
}

func logout(ctx *gin.Context) {
	var (
		idToken = ctx.Query("id_token")
	)

	if len(idToken) > 0 && len(idToken) < 1024 {
		tok, err := jwt.Parse(idToken, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return jwtPublicKey, nil
		})

		if err != nil {
			ctx.JSON(500, gin.H{
				"error":             "invalid_request",
				"error_description": "malformed jwt",
			})
			return
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims.Valid() != nil {
				ctx.JSON(400, errBadJWT)
				return
			}

			_, err := redisCluster.Del(ctx, claims["user"].(string)).Result()
			if err != nil {
				ctx.JSON(400, errBadJWT)
				return
			}

			ctx.Status(200)
		} else {
			ctx.JSON(400, errBadJWT)
		}
	} else {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid code",
		})
		return
	}
}

func getToken(ctx *gin.Context) {
	var (
		code         = ctx.Query("code")
		codeVerifier = ctx.Query("code_verifier")
	)

	if len(code) > 0 && len(code) < 1024 {
		tok, err := jwt.Parse(code, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return jwtPublicKey, nil
		})

		if err != nil {
			ctx.JSON(500, gin.H{
				"error":             "invalid_request",
				"error_description": "malformed code",
			})
			return
		}

		if _, ok := tok.Claims.(jwt.Claims); !ok && !tok.Valid {
			ctx.JSON(400, errBadJWT)
			return
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims.Valid() != nil {
				ctx.JSON(400, errBadJWT)
				return
			}

			if claims["sub"] != "authorize" {
				ctx.JSON(400, errBadJWT)
				return
			}

			val, err := redisCluster.LRange(ctx, claims["user"].(string), 0, -1).Result()
			if err != nil {
				ctx.JSON(400, errBadJWT)
			}
			if util.GetSHA256([]byte(codeVerifier)) != val[1] {
				ctx.JSON(400, gin.H{
					"error":             "access_denied",
					"error_description": "failed PKCE",
				})
				return
			}

			idJWT, err := createJWT(claims["user"].(string), 300, val[1], "id", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}
			refreshJWT, err := createJWT(claims["user"].(string), 360, val[1], "refresh", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}
			accessJWT, err := createJWT(claims["user"].(string), 300, val[1], "access", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}

			err = setTokens(ctx, claims["user"].(string), &tokenSet{id: idJWT, refresh: refreshJWT, access: accessJWT})
			if err != nil {
				ctx.JSON(500, errServiceConnection)
				return
			}

			ctx.JSON(200, gin.H{
				"access_token":  accessJWT,
				"refresh_token": refreshJWT,
				"id_token":      idJWT,
			})
		}
	} else {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid code",
		})
		return
	}
}


func refresh(ctx *gin.Context) {
	var (
		refreshToken = ctx.Query("refresh_token")
	)

	if len(refreshToken) > 0 && len(refreshToken) < 1024 {
		tok, err := jwt.Parse(refreshToken, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return jwtPublicKey, nil
		})
		if err != nil {
			ctx.JSON(500, gin.H{
				"error":             "invalid_request",
				"error_description": "malformed jwt",
			})
			return
		}
		if _, ok := tok.Claims.(jwt.Claims); !ok && !tok.Valid {
			ctx.JSON(400, errBadJWT)
			return
		}
		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims.Valid() != nil {
				ctx.JSON(400, errBadJWT)
				return
			}
			val, err := redisCluster.LRange(ctx, claims["user"].(string), 0, -1).Result()
			if err != nil {
				ctx.JSON(400, errBadJWT)
			}
			idJWT, err := createJWT(claims["user"].(string), 300, val[1], "id", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}
			refreshJWT, err := createJWT(claims["user"].(string), 360, val[1], "refresh", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}
			accessJWT, err := createJWT(claims["user"].(string), 300, val[1], "access", map[string]string{})
			if err != nil {
				ctx.JSON(500, errJWTCreation)
				return
			}

			_, errDel := redisCluster.Del(ctx, claims["user"].(string)).Result()
			if errDel != nil {
				ctx.JSON(400, errBadJWT)
				return
			}

			err = setTokens(ctx, claims["user"].(string), &tokenSet{ refresh: refreshJWT, access: accessJWT})
			if err != nil {
				ctx.JSON(500, errServiceConnection)
				return
			}

			ctx.JSON(200, gin.H{
				"access_token":  accessJWT,
				"refresh_token": refreshJWT,
				"id_token":      idJWT,
			})
		}
	} else {
		ctx.JSON(400, gin.H{
			"error":             "invalid_request",
			"error_description": "invalid jwt",
		})
		return
	}
}

func setTokens(ctx *gin.Context, userId string, tokens *tokenSet) error {
	res := redisCluster.SetEX(ctx, userId, tokens.id, time.Minute*5)
	if res.Err() != nil {
		return res.Err()
	}
	res = redisCluster.SetEX(ctx, userId, tokens.access, time.Minute*5)
	if res.Err() != nil {
		return res.Err()
	}
	res = redisCluster.SetEX(ctx, userId, tokens.refresh, time.Minute*6)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func createJWT(user string, expireIn int64, scope string, subject string, data map[string]string) (string, error) {
	now := time.Now().UTC()
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user":  user,
		"data":  data,
		"scope": scope,
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"sub":   subject,
		"exp":   now.Add(time.Second * time.Duration(expireIn)).Unix(),
	}).SignedString(jwtPrivateKey)

	if err != nil {
		return "", err
	}
	return token, nil
}

func verifyHash(password string, hash string) bool {
	p, salt, plainHash, err := decodeHash(hash)
	if err != nil {
		return false
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(plainHash, otherHash) == 1 {
		return true
	}
	return false
}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
