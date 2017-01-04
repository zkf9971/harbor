package daomongo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// AddProjectMember inserts a record to table project_member
func AddProjectMember(projectID bson.ObjectId, userID bson.ObjectId, role int) error {

	pmQ := bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}
	update := bson.M{
		"project_id":    projectID,
		"user_id":       userID,
		"role":          role,
		"creation_time": time.Now(),
		"update_time":   time.Now(),
	}

	_, _, err := mongo.UpsertDocument(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, update)
	return err
}

// UpdateProjectMember updates the record in table project_member
func UpdateProjectMember(projectID bson.ObjectId, userID bson.ObjectId, role int) error {

	pmQ := bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"role": role,
			},
		},
		ReturnNew: true,
	}

	_, _, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, change)
	return err
}

// DeleteProjectMember delete the record from table project_member
func DeleteProjectMember(projectID bson.ObjectId, userID bson.ObjectId) error {

	pmQ := bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}

	return mongo.RemoveDocument(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ)
}

// GetUserByProject gets all members of the project.
func GetUserByProject(projectID bson.ObjectId, queryUser models.User) ([]models.User, error) {

	// sql := `select u.user_id, u.username, r.name rolename, r.role_id as role
	// 	from user u
	// 	join project_member pm
	// 	on pm.project_id = ? and u.user_id = pm.user_id
	// 	join role r
	// 	on pm.role = r.role_id
	// 	where u.deleted = 0`

	//1: get project_member for user_id
	pmQ := bson.M{
		"project_id": projectID,
	}
	pMembers := []bson.M{}
	err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return nil, err
	}
	if len(pMembers) == 0 {
		return nil, fmt.Errorf("Project Member not found for user %s and project %v", queryUser.Username, projectID)
	}

	userIDs := []bson.ObjectId{}
	roleIDByUserID := map[string]int{}
	for _, pMember := range pMembers {
		userID := pMember["user_id"].(bson.ObjectId)
		userIDs = append(userIDs, userID)
		i, _ := strconv.Atoi(fmt.Sprintf("%v", pMember["role"]))
		roleIDByUserID[userID.String()] = i
	}

	//2: get user
	userQ := bson.M{
		"_id": bson.M{
			"$in": userIDs,
		},
	}
	if queryUser.Username != "" {
		userQ["username"] = bson.RegEx{queryUser.Username, "i"}
	}
	users := []models.User{}
	err = mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, []string{"_id"}, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("User %s not found", queryUser.Username)
	}

	//3: get role
	roleQ := bson.M{}
	roles := []models.Role{}
	err = mongo.FindRoles(mongo.REG_AUTH_ROLE_COLL, roleQ, nil, nil, &roles)
	if err != nil {
		return nil, err
	}

	roleByID := map[int]models.Role{}
	for _, role := range roles {
		roleByID[role.RoleID] = role
	}

	for i, user := range users {
		if role, ok := roleByID[roleIDByUserID[user.UserID.String()]]; ok {
			users[i].Role = role.RoleID
			users[i].Rolename = role.Name
		}
	}
	return users, err
}
