package daomongo

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"github.com/vmware/harbor/src/common/utils"
)

// Register is used for user to register, the password is encrypted before the record is inserted into database.
func Register(user models.User) (models.User, error) {

	userQ := bson.M{
		"username": user.Username,
	}

	salt := utils.GenerateRandomString()
	now := time.Now()
	update := bson.M{
		"username":      user.Username,
		"password":      utils.Encrypt(user.Password, salt),
		"realname":      user.Realname,
		"email":         user.Email,
		"comment":       user.Comment,
		"deleted":       user.Deleted,
		"salt":          salt,
		"sysadmin_flag": user.HasAdminRole,
		"creation_time": now,
		"update_time":   now,
	}

	user, _, err := mongo.UpsertUser(mongo.REG_AUTH_USER_COLL, userQ, update)
	if err != nil {
		return user, err
	}

	//userID := info.UpsertedId.(bson.ObjectId)
	return user, nil
}

// UserExists returns whether a user exists according username or Email.
func UserExists(user models.User, target string) (bool, error) {

	if user.Username == "" && user.Email == "" {
		return false, errors.New("user name and email are blank")
	}

	userQ := bson.M{}

	switch target {
	case "username":
		userQ["username"] = user.Username
	case "email":
		userQ["email"] = user.Email
	}

	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return false, err
	} else if len(users) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
