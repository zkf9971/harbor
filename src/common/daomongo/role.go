package daomongo

import (
	"fmt"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"gopkg.in/mgo.v2/bson"
)

// GetUserProjectRoles returns roles that the user has according to the project.
func GetUserProjectRoles(userID bson.ObjectId, projectID bson.ObjectId) ([]models.Role, error) {

	// sql := `select *
	// 	from role
	// 	where role_id =
	// 		(
	// 			select role
	// 			from project_member
	// 			where project_id = ? and user_id = ?
	// 		)`
	//1: get project_member
	pmQ := bson.M{
		"user_id":    userID,
		"project_id": projectID,
	}
	pMembers := []bson.M{}
	err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return nil, err
	}
	roleIds := []int{}
	for _, pMember := range pMembers {
		roleIds = append(roleIds, pMember["role"].(int))
	}
	//2: get role
	roleQ := bson.M{
		"role_id": bson.M{"$in": roleIds},
	}
	roles := []models.Role{}
	err = mongo.FindRoles(mongo.REG_AUTH_ROLE_COLL, roleQ, nil, nil, &roles)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// IsAdminRole returns whether the user is admin.
func IsAdminRole(userIDOrUsername interface{}) (bool, error) {
	u := models.User{}

	// log.Debugf("---- userIDOrUsername: %s %v", userIDOrUsername, bson.IsObjectIdHex(userIDOrUsername))

	// if bson.IsObjectIdHex(userIDOrUsername) {
	// 	u.UserID = bson.ObjectIdHex(userIDOrUsername)
	// } else {
	// 	u.Username = userIDOrUsername
	// }
	switch v := userIDOrUsername.(type) {
	case bson.ObjectId:
		u.UserID = v
	case string:
		u.Username = v
	default:
		return false, fmt.Errorf("invalid parameter, only bson.ObjectId and string are supported: %v", userIDOrUsername)
	}

	if u.UserID == "" && len(u.Username) == 0 {
		return false, nil
	}

	user, err := GetUser(u)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.HasAdminRole == 1, nil
}

// GetRoleByID ...
func GetRoleByID(id int) (*models.Role, error) {

	roleQ := bson.M{
		"role_id": id,
	}
	roles := []models.Role{}
	err := mongo.FindRoles(mongo.REG_AUTH_ROLE_COLL, roleQ, nil, nil, &roles)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, nil
	}

	return &roles[0], nil
}
