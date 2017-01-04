package daomongo

import (
	"fmt"
	"time"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/mongo"
	"gopkg.in/mgo.v2/bson"
)

// AddRepository adds a repo to the database.
func AddRepository(repo models.RepoRecord) error {

	// sql := "insert into repository (owner_id, project_id, name, description, pull_count, star_count, creation_time, update_time) " +
	// 	"select (select user_id as owner_id from user where username=?), " +
	// 	"(select project_id as project_id from project where name=?), ?, ?, ?, ?, ?, NULL "

	//get user_id
	userQ := bson.M{
		"username": repo.OwnerName,
	}
	users := []models.User{}
	err := mongo.FindUsers(mongo.REG_AUTH_USER_COLL, userQ, nil, nil, &users)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("User %s not found", repo.OwnerName)
	}
	user := users[0]
	//2: get project_id
	projectQ := bson.M{
		"name": repo.ProjectName,
	}
	projects := []models.Project{}
	err = mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return fmt.Errorf("Project %s not found", repo.ProjectName)
	}
	project := projects[0]
	//3: add repo
	repoQ := bson.M{
		"owner_id":   user.UserID,
		"project_id": project.ProjectID,
		"name":       repo.Name,
	}
	update := bson.M{
		"owner_id":      user.UserID,
		"project_id":    project.ProjectID,
		"name":          repo.Name,
		"description":   repo.Description,
		"pull_count":    repo.PullCount,
		"star_count":    repo.StarCount,
		"creation_time": time.Now(),
		"update_time":   time.Now(),
	}

	_, _, err = mongo.UpsertDocument(mongo.REG_AUTH_REPO_COLL, repoQ, update)
	return err
}

// GetRepositoryByName ...
func GetRepositoryByName(name string) (*models.RepoRecord, error) {

	repoQ := bson.M{
		"name": name,
	}
	repos := []models.RepoRecord{}
	err := mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil, &repos)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, nil
	}

	return &repos[0], err
}

// GetAllRepositories ...
func GetAllRepositories() ([]models.RepoRecord, error) {
	repoQ := bson.M{}
	repos := []models.RepoRecord{}
	err := mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil, &repos)

	return repos, err

}

// DeleteRepository ...
func DeleteRepository(name string) error {

	repoQ := bson.M{
		"name": name,
	}
	return mongo.RemoveDocument(mongo.REG_AUTH_REPO_COLL, repoQ)
}

// UpdateRepository ...
func UpdateRepository(repo models.RepoRecord) error {

	repoQ := bson.M{
		"_id": repo.RepositoryID,
	}

	repo.UpdateTime = time.Now()
	_, _, err := mongo.UpsertDocument(mongo.REG_AUTH_REPO_COLL, repoQ, repo)
	return err
}

// IncreasePullCount ...
func IncreasePullCount(name string) (err error) {

	repoQ := bson.M{
		"name": name,
	}
	repos := []models.RepoRecord{}
	err = mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil, &repos)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("repo %s not found", name)
	}
	repo := repos[0]

	update := bson.M{
		"$set": bson.M{
			"pull_count":  repo.PullCount + 1,
			"update_time": time.Now(),
		},
	}

	_, info, err := mongo.UpsertDocument(mongo.REG_AUTH_REPO_COLL, repoQ, update)
	if err != nil {
		return err
	}

	if info.Updated == 0 {
		err = fmt.Errorf("Failed to increase repository pull count with name: %s %s", name, err.Error())
	}
	return err
}

//RepositoryExists returns whether the repository exists according to its name.
func RepositoryExists(name string) bool {

	repoQ := bson.M{
		"name": name,
	}
	repos := []models.RepoRecord{}
	err := mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil, &repos)
	if err != nil {
		fmt.Printf("Failed to find repo %s. %v", name, err)
		return false
	}
	if len(repos) == 0 {
		return false
	}
	return true
}

// GetRepositoryByProjectName ...
func GetRepositoryByProjectName(name string) ([]models.RepoRecord, error) {

	projectQ := bson.M{
		"name": name,
	}
	projects := []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("project %s not found", name)
	}
	project := projects[0]

	repoQ := bson.M{
		"project_id": project.ProjectID,
	}
	repos := []models.RepoRecord{}
	err = mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil, &repos)
	return repos, err
}

//GetTopRepos returns the most popular repositories
func GetTopRepos(count int) ([]models.TopRepo, error) {
	topRepos := []models.TopRepo{}

	repositories := []models.RepoRecord{}

	repoQ := bson.M{}
	sort := []string{"-PullCount", "Name"}
	pagingOpts := map[string]int{
		"limit": count,
	}

	err := mongo.FindRepos(mongo.REG_AUTH_REPO_COLL, repoQ, sort, pagingOpts, &repositories)
	if err != nil {
		return nil, err
	}

	for _, repository := range repositories {
		topRepos = append(topRepos, models.TopRepo{
			RepoName:    repository.Name,
			AccessCount: repository.PullCount,
		})
	}

	return topRepos, nil
}

// GetTotalOfRepositories ...
func GetTotalOfRepositories(name string) (int64, error) {

	repoQ := bson.M{}
	if len(name) != 0 {
		repoQ["name"] = bson.RegEx{name, "i"}
	}

	n, err := mongo.CountDocuments(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil)
	if err != nil {
		return 0, err
	}
	return int64(n), err
}

// GetTotalOfPublicRepositories ...
func GetTotalOfPublicRepositories(name string) (int64, error) {

	projectQ := bson.M{
		"public": 1,
	}
	projects := []models.Project{}
	err := mongo.FindProjects(mongo.REG_AUTH_PROJECT_COLL, projectQ, nil, nil, &projects)
	if err != nil {
		return 0, err
	}
	if len(projects) == 0 {
		return 0, nil
	}
	publicProjectIds := []bson.ObjectId{}
	for _, project := range projects {
		publicProjectIds = append(publicProjectIds, project.ProjectID)
	}
	//2: count repos
	repoQ := bson.M{
		"project_id": bson.M{"$in": publicProjectIds},
	}
	if len(name) != 0 {
		repoQ["name"] = bson.RegEx{name, "i"}
	}
	n, err := mongo.CountDocuments(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil)
	if err != nil {
		return 0, err
	}
	return int64(n), err
}

// GetTotalOfUserRelevantRepositories ...
func GetTotalOfUserRelevantRepositories(userID bson.ObjectId, name string) (int64, error) {

	// sql := `select count(*)
	// 	from repository r
	// 	join (
	// 		select p.project_id, p.public
	// 			from project p
	// 			join project_member pm
	// 			on p.project_id = pm.project_id
	// 			where pm.user_id = ?
	// 	) as pp
	// 	on r.project_id = pp.project_id `
	pmQ := bson.M{
		"user_id": userID,
	}
	pMembers := []bson.M{}
	err := mongo.FindDocuments(mongo.REG_AUTH_PROJ_MEMBER_COLL, pmQ, nil, nil, &pMembers)
	if err != nil {
		return 0, err
	}
	if len(pMembers) == 0 {
		return 0, nil
	}
	userProjectIds := []bson.ObjectId{}
	for _, pMember := range pMembers {
		userProjectIds = append(userProjectIds, pMember["project_id"].(bson.ObjectId))
	}

	repoQ := bson.M{
		"_id": bson.M{"$in": userProjectIds},
	}
	if len(name) != 0 {
		repoQ["name"] = bson.RegEx{name, "i"}
	}
	n, err := mongo.CountDocuments(mongo.REG_AUTH_REPO_COLL, repoQ, nil, nil)
	if err != nil {
		return 0, err
	}
	return int64(n), err
}
