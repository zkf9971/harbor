package daomongo

import (
	"errors"
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"github.com/vmware/harbor/src/common/utils"

	"github.com/vmware/harbor/src/common/utils/log"
)

// GetUser ...
func GetUser(query models.User) (*models.User, error) {

	userQ := bson.M{}

	if query.UserID != "" {
		userQ["_id"] = query.UserID
	}

	if query.Username != "" {
		userQ["username"] = query.Username
	}

	if query.ResetUUID != "" {
		userQ["reset_uuid"] = query.ResetUUID
	}

	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)

	// log.Debugf("Users: %v", users)
	log.Debugf("Users len: %v", len(users))

	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

// LoginByDb is used for user to login with database auth mode.
func LoginByDb(auth models.AuthModel) (*models.User, error) {

	userQ := bson.M{
		"deleted": 0,
		"$or": []interface{}{
			bson.M{"username": auth.Principal},
			bson.M{"email": auth.Principal},
		},
	}

	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}

	user := users[0]

	if user.Password != utils.Encrypt(auth.Password, user.Salt) {
		return nil, nil
	}

	user.Password = "" //do not return the password

	return &user, nil
}

// ListUsers lists all users according to different conditions.
func ListUsers(query models.User) ([]models.User, error) {

	userQ := bson.M{
		"deleted": 0,
		"username": bson.M{
			"$ne": "admin",
		},
	}

	if query.Username != "" {
		userQ["username"] = bson.RegEx{query.Username, "i"}
	}

	sort := []string{"-_id"}

	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, sort, nil, &users)

	return users, err
}

// ToggleUserAdminRole gives a user admin role.
func ToggleUserAdminRole(userID bson.ObjectId, hasAdmin int) error {

	userQ := bson.M{
		"_id": userID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"sysadmin_flag": hasAdmin,
			},
		},
		ReturnNew: true,
	}

	_, info, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	if err != nil {
		return err
	}
	c := info.Updated
	if c == 0 {
		return errors.New("no record has been modified, toggle admin failed")
	}
	return nil
}

// ChangeUserPassword ...
func ChangeUserPassword(u models.User, oldPassword ...string) (err error) {
	if len(oldPassword) > 1 {
		return errors.New("wrong numbers of params")
	}

	userQ := bson.M{
		"_id": u.UserID,
	}
	//In some cases, it may no need to check old password, just as Linux change password policies.
	if len(oldPassword) != 0 {
		userQ["password"] = utils.Encrypt(oldPassword[0], u.Salt)
	}

	salt := utils.GenerateRandomString()
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"password": utils.Encrypt(u.Password, salt),
				"salt":     salt,
			},
		},
		ReturnNew: true,
	}

	_, info, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	if err != nil {
		return err
	}
	c := info.Updated
	if c == 0 {
		return errors.New("no record has been modified, change password failed")
	}
	return nil
}

// ResetUserPassword ...
func ResetUserPassword(u models.User) error {

	userQ := bson.M{
		"reset_uuid": u.ResetUUID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"password":   utils.Encrypt(u.Password, u.Salt),
				"reset_uuid": "",
			},
		},
		ReturnNew: true,
	}

	_, info, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	if err != nil {
		return err
	}
	count := info.Updated
	if count == 0 {
		return errors.New("no record be changed, reset password failed")
	}
	return nil
}

// UpdateUserResetUUID ...
func UpdateUserResetUUID(u models.User) error {

	userQ := bson.M{
		"email": u.Email,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"reset_uuid": u.ResetUUID,
			},
		},
		ReturnNew: true,
	}

	_, _, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	return err
}

// CheckUserPassword checks whether the password is correct.
func CheckUserPassword(query models.User) (*models.User, error) {

	currentUser, err := GetUser(query)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, nil
	}

	userQ := bson.M{
		"deleted":  0,
		"username": currentUser.Username,
		"password": utils.Encrypt(query.Password, currentUser.Salt),
	}

	users := []models.User{}
	err = mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		log.Warning("User principal does not match password. Current:", currentUser)
		return nil, nil
	}

	return &users[0], nil
}

// DeleteUser ...
func DeleteUser(userID bson.ObjectId) error {

	user, err := GetUser(models.User{
		UserID: userID,
	})
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s#%v", user.Username, user.UserID)
	email := fmt.Sprintf("%s#%v", user.Email, user.UserID)

	userQ := bson.M{
		"_id": userID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"deleted":  1,
				"username": name,
				"email":    email,
			},
		},
		ReturnNew: true,
	}

	_, _, err = mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	return err
}

// ChangeUserProfile ...
func ChangeUserProfile(user models.User) error {

	userQ := bson.M{
		"_id": user.UserID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"email":    user.Email,
				"realname": user.Realname,
				"comment":  user.Comment,
			},
		},
		ReturnNew: true,
	}

	_, _, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_USER_COLL, userQ, change)
	return err
}
