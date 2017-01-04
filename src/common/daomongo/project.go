package daomongo

import (
	"github.com/vmware/harbor/src/common/models"

	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/mongo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/vmware/harbor/src/common/utils/log"
)

//TODO:transaction, return err

// AddProject adds a project to the database along with project roles information and access log records.
func AddProject(project models.Project) (bson.ObjectId, error) {

	projectQ := bson.M{
		"name": project.Name,
	}
	now := time.Now()
	update := bson.M{
		"owner_id":      project.OwnerID,
		"name":          project.Name,
		"creation_time": now,
		"update_time":   now,
		"deleted":       project.Deleted,
		"public":        project.Public,
	}

	_, info, err := mongo.UpsertDocument(mongo.REG_AUTH_PROJECT_COLL, projectQ, update)
	if err != nil {
		return "", err
	}

	projectID := info.UpsertedId.(bson.ObjectId)

	if err = AddProjectMember(projectID, project.OwnerID, models.PROJECTADMIN); err != nil {
		return projectID, err
	}

	accessLog := models.AccessLog{UserID: project.OwnerID, Username: project.OwnerName, ProjectID: projectID, RepoName: project.Name + "/", RepoTag: "N/A", GUID: "N/A", Operation: "create", OpTime: time.Now()}
	err = AddAccessLog(accessLog)

	return projectID, err
}

// IsProjectPublic ...
func IsProjectPublic(projectName string) bool {
	project, err := GetProjectByName(projectName)
	if err != nil {
		log.Errorf("Error occurred in GetProjectByName: %v", err)
		return false
	}
	if project == nil {
		return false
	}
	return project.Public == 1
}

//ProjectExists returns whether the project exists according to its name of ID.
func ProjectExists(nameOrID string) (bool, error) {

	projectQ := bson.M{
		"deleted": 0,
	}

	if bson.IsObjectIdHex(nameOrID) {
		projectQ["_id"] = nameOrID
	} else {
		projectQ["name"] = nameOrID
	}

	projects := []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return false, err
	}

	return len(projects) > 0, nil
}

// GetProjectByID ...
func GetProjectByID(id bson.ObjectId) (*models.Project, error) {

	projectQ := bson.M{
		"deleted": 0,
		"_id":     id,
	}

	projects := []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	return &projects[0], nil
}

// GetProjectByName ...
func GetProjectByName(name string) (*models.Project, error) {

	projectQ := bson.M{
		"deleted": 0,
		"name":    name,
	}

	projects := []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	return &projects[0], nil
}

// GetPermission gets roles that the user has according to the project.
func GetPermission(username, projectName string) (string, error) {

	// sql := `select r.role_code from role as r
	// 	inner join project_member as pm on r.role_id = pm.role
	// 	inner join user as u on u.user_id = pm.user_id
	// 	inner join project p on p.project_id = pm.project_id
	// 	where u.username = ? and p.name = ? and u.deleted = 0 and p.deleted = 0`
	//1: get user_id
	userQ := bson.M{
		"username": username,
	}
	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("User %s not found", username)
	}
	user := users[0]

	//2: get project_id
	projectQ := bson.M{
		"name": projectName,
	}
	projects := []models.Project{}
	err = mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return "", err
	}
	if len(projects) == 0 {
		return "", fmt.Errorf("Project %s not found", projectName)
	}
	project := projects[0]

	//3: get project_member document based on user_id and project_id
	pmQ := bson.M{
		"user_id":    user.UserID,
		"project_id": project.ProjectID,
	}
	pMembers := []bson.M{}
	err = mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return "", err
	}
	if len(pMembers) == 0 {
		return "", fmt.Errorf("Project Member not found for user %s and project %s", username, projectName)
	}
	pMember := pMembers[0]
	//4: get role
	roleQ := bson.M{
		"role_id": pMember["role"],
	}
	roles := []models.Role{}
	err = mongo.FindRoles(mongo.REG_AUTH_ROLE_COLL, roleQ, nil, nil, &roles)
	if err != nil {
		return "", err
	}
	if len(roles) == 0 {
		return "", fmt.Errorf("Role %s not found", pMember["role"])
	}
	role := roles[0]

	return role.RoleCode, nil
}

// ToggleProjectPublicity toggles the publicity of the project.
func ToggleProjectPublicity(projectID bson.ObjectId, publicity int) error {

	projectQ := bson.M{
		"_id": projectID,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"public": publicity,
			},
		},
		ReturnNew: true,
	}

	_, _, err := mongo.FindAndModifyDocument(mongo.REG_AUTH_PROJECT_COLL, projectQ, change)
	return err
}

// SearchProjects returns a project list,
// which satisfies the following conditions:
// 1. the project is not deleted
// 2. the project is public or the user is a member of the project
func SearchProjects(userID bson.ObjectId) ([]models.Project, error) {

	pmQ := bson.M{
		"user_id": userID,
	}
	pMembers := []bson.M{}
	err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return nil, err
	}
	userProjectIds := []bson.ObjectId{}
	for _, pMember := range pMembers {
		userProjectIds = append(userProjectIds, pMember["project_id"].(bson.ObjectId))
	}

	// sql := `select distinct p.project_id, p.name, p.public
	// 	from project p
	// 	left join project_member pm on p.project_id = pm.project_id
	// 	where (pm.user_id = ? or p.public = 1) and p.deleted = 0`
	projectQ := bson.M{
		"$or": []interface{}{
			bson.M{
				"_id": bson.M{"$in": userProjectIds},
			},
			bson.M{
				"public": 1,
			},
		},
		"deleted": 0,
	}

	projects := []models.Project{}
	err = mongo.FindDistinctProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, "name", &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

//GetTotalOfUserRelevantProjects returns the total count of
// user relevant projects
func GetTotalOfUserRelevantProjects(userID bson.ObjectId, projectName string) (int64, error) {

	pmQ := bson.M{
		"user_id": userID,
	}
	pMembers := []bson.M{}
	err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return 0, err
	}
	userProjectIds := []bson.ObjectId{}
	for _, pMember := range pMembers {
		userProjectIds = append(userProjectIds, pMember["project_id"].(bson.ObjectId))
	}

	// sql := `select count(*) from project p
	// 		left join project_member pm
	// 		on p.project_id = pm.project_id
	//  		where p.deleted = 0 and pm.user_id= ?`
	projectQ := bson.M{
		"deleted": 0,
		"_id":     bson.M{"$in": userProjectIds},
	}
	if projectName != "" {
		projectQ["name"] = bson.RegEx{projectName, "i"}
	}

	projects := []models.Project{}
	err = mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return 0, err
	}

	return int64(len(projects)), nil
}

// GetUserRelevantProjects returns the user relevant projects
// args[0]: public, args[1]: limit, args[2]: offset
func GetUserRelevantProjects(userID bson.ObjectId, projectName string, args ...int64) ([]models.Project, error) {
	return getProjects(userID, projectName, args...)
}

// GetTotalOfProjects returns the total count of projects
func GetTotalOfProjects(name string, public ...int) (int64, error) {

	projectQ := bson.M{
		"deleted": 0,
	}
	if len(name) > 0 {
		projectQ["name"] = name
	}

	if len(public) > 0 {
		projectQ["public"] = public[0]
	}

	n, err := mongo.CountDocuments(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil)
	if err != nil {
		return 0, err
	}
	return int64(n), err
}

// GetProjects returns project list
// args[0]: public, args[1]: limit, args[2]: offset
func GetProjects(name string, args ...int64) ([]models.Project, error) {
	return getProjects("", name, args...)
}

func getProjects(userID bson.ObjectId, name string, args ...int64) ([]models.Project, error) {
	projects := []models.Project{}

	projectQ := bson.M{
		"deleted": 0,
	}

	if userID != "" { //get user's projects
		// sql = `select distinct p.project_id, p.owner_id, p.name,
		// 			p.creation_time, p.update_time, p.public, pm.role role
		// 	from project p
		// 	left join project_member pm
		// 	on p.project_id = pm.project_id
		// 	where p.deleted = 0 and pm.user_id= ?`
		pmQ := bson.M{
			"user_id": userID,
		}
		pMembers := []bson.M{}
		err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
		if err != nil {
			return nil, err
		}
		userProjectIds := []bson.ObjectId{}
		for _, pMember := range pMembers {
			userProjectIds = append(userProjectIds, pMember["project_id"].(bson.ObjectId))
		}

		projectQ["_id"] = bson.M{"$in": userProjectIds}
	}

	if name != "" {
		projectQ["name"] = bson.RegEx{name, "i"}
	}

	var pagingOpts map[string]int

	switch len(args) {
	case 1:
		projectQ["public"] = args[0]
	case 2:
		pagingOpts = map[string]int{}
		pagingOpts["limit"] = int(args[0])
		pagingOpts["offset"] = int(args[1])
	case 3:
		projectQ["public"] = args[0]
		pagingOpts = map[string]int{}
		pagingOpts["limit"] = int(args[1])
		pagingOpts["offset"] = int(args[2])
	}
	sort := []string{"name"}

	projects = []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, sort, pagingOpts, &projects)

	return projects, err
}

// DeleteProject ...
func DeleteProject(id bson.ObjectId) error {
	project, err := GetProjectByID(id)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s#%v", project.Name, project.ProjectID)

	projectQ := bson.M{
		"_id": id,
	}
	change := mgo.Change{
		Update: bson.M{
			"$set": bson.M{
				"deleted": 1,
				"name":    name,
			},
		},
		ReturnNew: true,
	}

	_, _, err = mongo.FindAndModifyDocument(mongo.REG_AUTH_PROJECT_COLL, projectQ, change)
	return err
}
