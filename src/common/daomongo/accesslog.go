package daomongo

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"fmt"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"github.com/vmware/harbor/src/common/utils/log"
)

// AddAccessLog persists the access logs
func AddAccessLog(accessLog models.AccessLog) error {

	return mongo.InsertDocument(mongo.REG_AUTH_ACCESSLOG_COLL, accessLog)
}

// GetTotalOfAccessLogs ...
func GetTotalOfAccessLogs(query models.AccessLog) (int64, error) {

	logQ := bson.M{
		"project_id": query.ProjectID,
	}

	if query.Username != "" {
		userQ := bson.M{
			"username": bson.RegEx{query.Username, "i"},
		}
		users := []models.User{}
		err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
		if err != nil {
			return 0, err
		}
		userIDs := []bson.ObjectId{}
		for _, user := range users {
			userIDs = append(userIDs, user.UserID)
		}
		logQ["user_id"] = bson.M{
			"$in": userIDs,
		}
	}

	genFilterClauses(query, &logQ)

	total, err := mongo.CountDocuments(mongo.REG_AUTH_ACCESSLOG_COLL, logQ, nil, nil)
	if err != nil {
		return 0, err
	}
	return int64(total), nil
}

//GetAccessLogs gets access logs according to different conditions
func GetAccessLogs(query models.AccessLog, limit, offset int64) ([]models.AccessLog, error) {

	logQ := bson.M{
		"project_id": query.ProjectID,
	}

	if query.Username != "" {
		userQ := bson.M{
			"username": bson.RegEx{query.Username, "i"},
		}
		users := []models.User{}
		err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
		if err != nil {
			return nil, err
		}
		userIDs := []bson.ObjectId{}
		for _, user := range users {
			userIDs = append(userIDs, user.UserID)
		}
		logQ["user_id"] = bson.M{
			"$in": userIDs,
		}
	}

	genFilterClauses(query, &logQ)

	sort := []string{"-op_time"}
	pagingOpts := map[string]int{
		"limit":  int(limit),
		"offset": int(offset),
	}

	logs := []models.AccessLog{}
	err := mongo.FindAccessLogs(mongo.REG_AUTH_ACCESSLOG_COLL, logQ, sort, pagingOpts, &logs)
	//TODO: map user_id to username
	return logs, err
}

func genFilterClauses(query models.AccessLog, logQp *bson.M) {

	logQ := *logQp

	if query.Operation != "" {
		logQ["operation"] = query.Operation
	}
	if query.RepoName != "" {
		logQ["repo_name"] = query.RepoName
	}
	if query.RepoTag != "" {
		logQ["repo_tag"] = query.RepoTag
	}
	if query.Keywords != "" {
		keywordArray := []string{}
		keywordList := strings.Split(query.Keywords, "/")
		num := len(keywordList)
		for i := 0; i < num; i++ {
			if keywordList[i] != "" {
				keywordArray = append(keywordArray, keywordList[i])
			}
		}
		logQ["operation"] = bson.M{
			"$in": keywordArray,
		}
	}
	if query.BeginTimestamp > 0 {
		logQ["op_time"] = bson.M{
			"$gte": query.BeginTime,
		}
	}
	if query.EndTimestamp > 0 {
		logQ["op_time"] = bson.M{
			"$lte": query.EndTime,
		}
	}

	return
}

// AccessLog ...
func AccessLog(username, projectName, repoName, repoTag, action string) error {

	userQ := bson.M{
		"username": username,
	}
	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("user %s not found", username)
	}
	user := users[0]

	projectQ := bson.M{
		"name": projectName,
	}
	projects := []models.Project{}
	err = mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("project %s not found", projectName)
	}
	project := projects[0]

	accessLog := models.AccessLog{
		UserID:    user.UserID,
		Username:  user.Username,
		ProjectID: project.ProjectID,
		RepoName:  repoName,
		RepoTag:   repoTag,
		Operation: action,
		OpTime:    time.Now(),
	}

	err = mongo.InsertDocument(mongo.REG_AUTH_ACCESSLOG_COLL, accessLog)
	if err != nil {
		log.Errorf("error in AccessLog: %v ", err)
	}
	return err
}

//GetRecentLogs returns recent logs according to parameters
func GetRecentLogs(userID bson.ObjectId, linesNum int, startTime, endTime string) ([]models.AccessLog, error) {
	logs := []models.AccessLog{}

	isAdmin, err := IsAdminRole(userID)
	if err != nil {
		return logs, err
	}

	logQ := bson.M{}
	if !isAdmin {
		pmQ := bson.M{
			"user_id": userID,
		}
		pMembers := []bson.M{}
		err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
		if err != nil {
			return logs, err
		}
		projectIDs := []bson.ObjectId{}
		for _, pMember := range pMembers {
			projectIDs = append(projectIDs, pMember["project_id"].(bson.ObjectId))
		}
		logQ["project_id"] = bson.M{
			"$in": projectIDs,
		}
	}

	if startTime != "" {
		logQ["op_time"] = bson.M{
			"$gte": startTime,
		}
	}

	if endTime != "" {
		logQ["op_time"] = bson.M{
			"$lte": endTime,
		}
	}

	sort := []string{"-op_time"}
	pagingOpts := map[string]int{}
	if linesNum != 0 {
		pagingOpts["limit"] = linesNum
	}

	err = mongo.FindAccessLogs(mongo.REG_AUTH_ACCESSLOG_COLL, logQ, sort, pagingOpts, &logs)
	if err != nil {
		return logs, err
	}
	//TODO map user_id to username
	return logs, nil
}

// GetAccessLogCreator ...
func GetAccessLogCreator(repoName string) (string, error) {

	logQ := bson.M{
		"operation": "push",
		"repo_name": repoName,
	}
	sort := []string{"-op_time"}
	pagingOpts := map[string]int{
		"limit": 1,
	}
	accessLogs := []models.AccessLog{}
	err := mongo.FindAccessLogs(mongo.REG_AUTH_ACCESSLOG_COLL, logQ, sort, pagingOpts, &accessLogs)
	if err != nil {
		return "", err
	}
	if len(accessLogs) == 0 {
		return "", fmt.Errorf("access log not found by operation 'push' and repo_name '%s'", repoName)
	}
	accessLog := accessLogs[0]

	userQ := bson.M{
		"user_id": accessLog.UserID,
	}
	users := []models.User{}
	err = mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", fmt.Errorf("user not found by id %v", accessLog.UserID)
	}

	return users[0].Username, nil
}

// CountPull ...
func CountPull(repoName string) (int64, error) {

	logQ := bson.M{
		"repo_name": repoName,
		"operation": "pull",
	}

	num, err := mongo.CountDocuments(mongo.REG_AUTH_ACCESSLOG_COLL, logQ, nil, nil)
	if err != nil {
		log.Errorf("error in CountPull: %v ", err)
		return 0, err
	}
	return int64(num), nil
}
