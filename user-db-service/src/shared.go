package main

import (
	"database/sql"
	"encoding/base64"
	pb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	"github.com/open-exam/open-exam-backend/util"
	"golang.org/x/crypto/argon2"
)

func getUser(mode int, data string, getPass bool) (*pb.User, error) {
	var (
		user = &pb.User{}
		rows *sql.Rows
		err  error
	)

	if mode == 0 {
		rows, err = db.Query("SELECT * FROM users WHERE id=?", data)
	} else {
		rows, err = db.Query("SELECT * FROM users WHERE email=?", data)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		if getPass {
			if err := rows.Scan(&user.Id, &user.Email, &user.Type, &user.Password); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&user.Id, &user.Email, &user.Type); err != nil {
				return nil, err
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return user, nil
}

func generateFromPassword(password string) (encodedHash string, err error) {
	salt := util.GenerateRandomBytes(16)

	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return "$argon2id$v=19$m=65536,t=3,p=2$" + b64Salt + "$" + b64Hash, nil
}

func generatePassword(len uint32) string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	res := ""

	for i := uint32(0); i < len; i++ {
		res += string(charSet[util.GenerateRandomBytes(1)[0]%62])
	}

	return res
}
