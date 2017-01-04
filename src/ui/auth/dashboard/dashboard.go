package dashboard

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"

	"github.com/jeffail/gabs"
	"github.com/vmware/harbor/src/ui/auth"
)

const (
	host360         = "platform.appcelerator.com"
	authPath        = "/api/v1/auth/login"
	logoutPath      = "/api/v1/auth/logout"
	orgInfoPath     = "/api/v1/user/organizations"
	thisEnvAdminURL = "http://admin.cloudapp-1.appctest.com"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

/**
 * Authenticate user against appcelerator 360 (dashboard). This is for enterprise user only.
 * @param username
 * @param password
 * @param cb
 */
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	loginURL := "https://" + host360 + authPath
	log.Debugf("Login user %s using password against %s...", m.Principal, loginURL)

	username := m.Principal
	creds := url.Values{}
	creds.Set("username", username)
	creds.Add("password", m.Password)
	// v.Encode() == "name=Ava&friend=Jess&friend=Sarah&friend=Zoe"

	//curl -i -b cookies.txt -c cookies.txt -F "username=mgoff@appcelerator.com" -F "password=food" http://360-dev.appcelerator.com/api/v1/auth/login
	/*
	   response for bad username/password
	   HTTP/1.1 400 Bad Request
	   X-Powered-By: Express
	   Access-Control-Allow-Origin: *
	   Access-Control-Allow-Methods: GET, POST, DELETE, PUT
	   Access-Control-Allow-Headers: Content-Type, api_key
	   Content-Type: application/json; charset=utf-8
	   Content-Length: 79
	   Date: Fri, 19 Apr 2013 01:25:24 GMT
	   Connection: keep-alive

	   {"success":false,"description":"Invalid password.","code":400,"internalCode":2}
	*/
	resp, err := http.PostForm(loginURL, creds)

	if err != nil {
		log.Errorf("Failed to login to dashboard. %v", err)
		return nil, err
	}

	//log.Debugf("resp: %v", resp)

	if resp.StatusCode != 200 {
		log.Debugf("dashboard returns status %s", resp.Status)
		return nil, errors.New("authentication failed")
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response body. %v", err)
		return nil, err
	}

	jsonBody, err := gabs.ParseJSON(bodyBuf)
	if err != nil {
		log.Errorf("Failed to parse response body. %v", err)
		return nil, err
	}

	// 'set-cookie': ['t=UpnUzNztGWO7K8A%2BCYihZz056Bk%3D; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'un=mgoff%40appcelerator.com; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'sid=33f33a6b7f8fef7b0fc649654187d467; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'dvid=2019bea3-9e7b-48e3-890f-00e3e22b39e2; Path=/; Expires=Sat, 17 Oct 2015 06:27:19 GMT',
	// 	'connect.sid=s%3Aj0kX71OMFpIQ11Vf1ruhqJLH.on4RLy9q9tpVqnUeoQJBWlDPiB6bS8rWWhq8sOCDGPc; Domain=360-dev.appcelerator.com; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT; HttpOnly'
	// ]
	// {
	// 	"success": true,
	// 	"result": {
	// 		"success": true,
	// 		"username": "mgoff@appcelerator.com",
	// 		"email": "mgoff@appcelerator.com",
	// 		"guid": "ae6150453b3599b2875b311c40785b40",
	// 		"org_id": 14301,
	// 		"connect.sid": "s:QGW1cqj5h9B3fL6jwJTtjkuT.iuwQ23WOgiK/E+QfkRNVWi7G5S9DA00Li6BQPLGkROM"
	// }

	cookie := resp.Header.Get("set-cookie")
	if cookie == "" {
		log.Error("No cookie found in response")
		return nil, errors.New("authentication failed")
	}

	sid := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	log.Debugf("sid: %s", sid)

	// for _, cookie := range cookies.([]string) {
	// 	log.Debugf("cookie: %s", cookie)
	// }

	success := jsonBody.Path("success").Data().(bool)
	if !success {
		log.Error("dashboard returns false for success field")
		return nil, errors.New("authentication failed")
	}

	// user := bson.M{
	// 	"username": jsonBody.Path("result.username").Data().(string),
	// 	"email":    jsonBody.Path("result.email").Data().(string),
	// 	"guid":     jsonBody.Path("result.guid").Data().(string),
	// }
	// if jsonBody.Path("result.firstname").Data() != nil {
	// 	user["firstname"] = jsonBody.Path("result.firstname").Data().(string)
	// } else {
	// 	user["firstname"] = jsonBody.Path("result.username").Data().(string)
	// }

	mUser := &models.User{}
	return mUser, nil
}

//TOOD use request module to support proxy
func logoutFromDashboard(sid360 string) (err error) {

	log.Debugf("Logout session %s from Appcelerator 360.", sid360)

	logOutURL := "https://" + host360 + logoutPath

	client := &http.Client{}

	req, err := http.NewRequest("GET", logOutURL, nil)

	req.Header.Add("Cookie", "connect.sid="+sid360)

	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("Failed to logout from dashboard. %v", err)
		return
	}

	//log.Debugf("resp: %v", resp)

	if resp.StatusCode == 400 {
		log.Warning("Failed to logout from dashboard. Session is invalid")
		err = errors.New("failed to logout from dashboard. Session is invalid")
		return
	}

	if resp.StatusCode != 200 {
		log.Debugf("dashboard returns status %s", resp.Status)
		err = errors.New("Failed to logout from dashboard")
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response body. %v", err)
		return
	}

	jsonBody, err := gabs.ParseJSON(bodyBuf)
	if err != nil {
		log.Errorf("Failed to parse response body. %v", err)
		return
	}

	success := jsonBody.Path("success").Data().(bool)
	if !success {
		log.Error("dashboard returns false for success field")
		err = errors.New("Failed to logout from dashboard")
		return
	}

	return
}

func init() {
	auth.Register("dashboard", &Auth{})
}
