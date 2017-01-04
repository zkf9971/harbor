package arrowcloud

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jeffail/gabs"
	"gopkg.in/mgo.v2/bson"

	dao "github.com/vmware/harbor/src/common/daomongo"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

const (
	ARROW_CLOUD_SESSION_SECRET = "stratus"
	ARROW_CLOUD_DB             = "arrowcloud"
	ARROW_CLOUD_SESSION_COLL   = ARROW_CLOUD_DB + ":sessions"
	ARROW_CLOUD_USERS_COLL     = ARROW_CLOUD_DB + ":users"
	ARROW_CLOUD_APPS_COLL      = ARROW_CLOUD_DB + ":apps"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

/**
 * Authenticate user against appcelerator 360 (dashboard). This is for enterprise user only.
 * Password field is arrowcloud session Id.
 * We only need to verify the session is valid.
 * @param username
 * @param password
 * @param cb
 */
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	log.Debugf("Login user %s using arrowcloud session Id %s...", m.Principal, m.Password)

	// 1: get session from arrowcloud.session collection (m.Password is session Id)
	userID, err := getUserIDFromSession(m.Password)
	if err != nil {
		return nil, err
	}

	// 2: get user from arrowcloud.users collection
	arrowCloudUser, err := getUser(userID)
	if err != nil {
		return nil, err
	}

	log.Debugf("arrowcloud user: %+v", arrowCloudUser)

	// 3: add user to registry_auth.user if it has not been done
	regUser, err := registerUser(arrowCloudUser)
	if err != nil {
		return nil, err
	}

	// 4: create a project for each user's organization which has not corresponding project
	projectIDs, err := createProjectsForOrgs(arrowCloudUser)
	if err != nil {
		return nil, err
	}

	// 5: add user to the projects as developer
	err = addUserToProjects(projectIDs, regUser.UserID)

	return &regUser, err
}

func addUserToProjects(projectIDs []bson.ObjectId, userID bson.ObjectId) error {

	for _, projectID := range projectIDs {

		pmQ := bson.M{
			"project_id": projectID,
			"user_id":    userID,
		}
		pMembers := []bson.M{}
		err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
		if err != nil {
			return err
		}
		if len(pMembers) == 0 {
			err := dao.AddProjectMember(projectID, userID, models.DEVELOPER)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// {
// 	"_id" : ObjectId("58690bd71fe0c733b70d4cac"),
// 	"guid" : "e5fd038ab0ddf6bf04479e289e38141c",
// 	"email" : "yjin@appcelerator.com",
// 	"username" : "yjin@appcelerator.com",
// 	"firstname" : "Yuping",
// 	"orgs_360" : [
// 		{
// 			"id" : "14301",
// 			"name" : "appcelerator Inc.",
// 			"admin" : true,
// 			"node_acs_admin" : true
// 		}
// 	],
// 	"orgs_360_updated_at" : ISODate("2017-01-01T14:01:59.274Z")
// }
func createProjectsForOrgs(arrowCloudUser bson.M) (projectIDs []bson.ObjectId, err error) {

	//admin will be the owner of all projects (namespaces) for organizations
	userQ := bson.M{
		"username": "admin",
	}
	users := []models.User{}
	err = mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return
	}
	if len(users) == 0 {
		err = errors.New("admin user not found")
		return
	}
	admin := users[0]

	//all IDs of the projects corresponding to user's organizations
	projectIDs = []bson.ObjectId{}

	orgs := arrowCloudUser["orgs_360"].([]interface{})

	orgIDs := []string{}
	orgNameByID := map[string]string{}
	for _, orgI := range orgs {
		org := orgI.(bson.M)
		orgID := org["id"].(string)
		orgIDs = append(orgIDs, orgID)
		orgNameByID[orgID] = org["name"].(string)
	}
	//find all existing projects for user's organizations so that we don't create them again
	projectQ := bson.M{
		"name": bson.M{
			"$in": orgIDs,
		},
	}
	projects := []models.Project{}
	err = mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return
	}
	for _, project := range projects {
		delete(orgNameByID, project.Name) //project.Name is org ID
		projectIDs = append(projectIDs, project.ProjectID)
	}

	//create projects for organizations which have not been done before
	for orgID, orgName := range orgNameByID {
		project := models.Project{
			OwnerID:   admin.UserID,
			OwnerName: admin.Username,
			Name:      orgID,
			Comment:   orgName,
			Deleted:   0,
			Public:    0,
		}

		projectID, ierr := dao.AddProject(project)
		if ierr != nil {
			err = ierr
			return
		}
		projectIDs = append(projectIDs, projectID)
	}

	return
}

//registerUser adds user to registry_auth.user if it's not there
func registerUser(arrowCloudUser bson.M) (regUser models.User, err error) {

	userQ := bson.M{
		"email": arrowCloudUser["email"],
	}

	regUsers := []models.User{}
	err = mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &regUsers)
	if err != nil {
		return
	}
	if len(regUsers) > 0 { //already registered to registry_auth.user
		regUser = regUsers[0]
		return
	}

	regUser = models.User{
		Username:     arrowCloudUser["email"].(string),
		Email:        arrowCloudUser["email"].(string),
		Password:     "n/a",
		Realname:     arrowCloudUser["firstname"].(string),
		Comment:      "arrowcloud user registered automatically",
		HasAdminRole: 0,
		Deleted:      0,
	}

	regUser, err = dao.Register(regUser)
	return
}

func getUser(userID string) (bson.M, error) {

	userQ := bson.M{
		"_id": bson.ObjectIdHex(userID),
	}

	users := []bson.M{}
	err := mongo.FindDocuments(ARROW_CLOUD_USERS_COLL, userQ, nil, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found by ID %s", userID)
	}
	return users[0], nil
}

func getUserIDFromSession(sessionID string) (userID string, err error) {

	log.Debugf("get userID from session: %s", sessionID)
	//decode session ID cookie from base64 encoding
	decodedSID, err := deCodeCookie(sessionID)
	if err != nil {
		return "", err
	}

	sessObjID := parseSignedCookie(decodedSID, ARROW_CLOUD_SESSION_SECRET)

	sessionQ := bson.M{
		"_id": sessObjID,
		"$or": []interface{}{
			bson.M{
				"expires": bson.M{
					"$exists": false,
				},
			},
			bson.M{
				"expires": bson.M{
					"$gt": time.Now(),
				},
			},
		},
	}

	sessions := []bson.M{}
	err = mongo.FindDocuments(ARROW_CLOUD_SESSION_COLL, sessionQ, nil, nil, &sessions)
	if err != nil {
		return "", err
	}

	if len(sessions) == 0 {
		return "", fmt.Errorf("session %s not found", sessionID)
	}
	session := sessions[0]
	// {
	// 	"_id" : "M3fGPaOdPX/ri6aXw7Rs/uCD",
	// 	"session" : "{\"cookie\":{\"originalMaxAge\":1209600000,\"expires\":\"2017-01-15T14:01:59.356Z\",\"httpOnly\":true,\"path\":\"/\"},\"lastCommand\":\"list\",\"mid\":\"be19c3bc52dc83ec5a11559028503582cd5cd07a\",\"username\":\"yjin@appcelerator.com\",\"auth\":{\"userId\":\"58690bd71fe0c733b70d4cac\",\"loggedIn\":true},\"sid_360\":\"s%3ATlW7octQkxrYCfpy18NRrmdsYKdK-7lc.%2BVvSqYrmTFfAz2P%2Fy5AvdW6fiKkMuymzaFxD92w5zmI\"}",
	// 	"expires" : ISODate("2017-01-15T14:01:59.356Z")
	// }
	// parsed session field
	// {
	//   "cookie": {
	//     "originalMaxAge": 1209600000,
	//     "expires": "2017-01-15T14:01:59.356Z",
	//     "httpOnly": true,
	//     "path": "/"
	//   },
	//   "lastCommand": "list",
	//   "mid": "be19c3bc52dc83ec5a11559028503582cd5cd07a",
	//   "username": "yjin@appcelerator.com",
	//   "auth": {
	//     "userId": "58690bd71fe0c733b70d4cac",
	//     "loggedIn": true
	//   },
	//   "sid_360": "s%3ATlW7octQkxrYCfpy18NRrmdsYKdK-7lc.%2BVvSqYrmTFfAz2P%2Fy5AvdW6fiKkMuymzaFxD92w5zmI"
	//  }

	escapedSessionContent := session["session"].(string)
	log.Debugf("escaped session content: %s", escapedSessionContent)

	// unquoted, err := strconv.Unquote(escapedSessionContent)
	// if err != nil {
	// 	return "", err
	// }
	// log.Debugf("unquoted session content: %s", unquoted)

	sessionJSON, err := gabs.ParseJSON([]byte(escapedSessionContent))
	if err != nil {
		return "", err
	}

	userID, ok := sessionJSON.Path("auth.userId").Data().(string)
	if !ok {
		return "", errors.New("error parsing session for auth.userId")
	}

	return userID, nil
}

//connect.sid=s%3AM3fGPaOdPX%2Fri6aXw7Rs%2FuCD.yyni98IuUYOjqqZ%2FXzdJmoPf1DVylvMCe5%2B62feHAl8
func deCodeCookie(str string) (string, error) {
	log.Debugf("decode sessionID from URL encoding: %s", str)
	data, err := url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	log.Debugf("decoded sessionID: %s", data)
	return data, nil
}

//from nodejs connect/lib/utils.js
func parseSignedCookie(str, secret string) string {
	if strings.HasPrefix(str, "s:") {
		return unsign(str[2:], secret)
	}
	return str
}

// exports.parseSignedCookie = function(str, secret){
//   return 0 == str.indexOf('s:')
//     ? exports.unsign(str.slice(2), secret)
//     : str;
// };

//from nodejs connect/lib/utils.js
func unsign(val, secret string) string {
	log.Debugf("unsign %s with secret %s", val, secret)
	str := val[0:strings.LastIndex(val, ".")]
	log.Debugf("session ObjectId: %s", str)
	if val == sign(str, secret) {
		log.Debugf("verified successfully")
		return str
	}
	return ""
}

// exports.unsign = function(val, secret){
//   var str = val.slice(0, val.lastIndexOf('.'));
//   return exports.sign(str, secret) == val
//     ? str
//     : false;
// };

//from nodejs connect/lib/utils.js
func sign(val, secret string) string {

	log.Debugf("sign %s with %s", val, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(val))
	generatedMAC := mac.Sum(nil)

	log.Debugf("generated mac: %v", generatedMAC)
	base64Str := base64.StdEncoding.EncodeToString(generatedMAC)

	log.Debugf("base64 encoding: %s", base64Str)
	equalSigns := regexp.MustCompile(`=+$`)

	equalSignsRemoved := equalSigns.ReplaceAllLiteralString(base64Str, "")
	log.Debugf("equal signs removed: %s", equalSignsRemoved)
	return val + "." + equalSignsRemoved
}

// exports.sign = function(val, secret){

//   return val + '.' + crypto
//     .createHmac('sha256', secret)
//     .update(val)
//     .digest('base64')
//     .replace(/=+$/, '');
// };

func init() {
	auth.Register("arrowcloud", &Auth{})
}
