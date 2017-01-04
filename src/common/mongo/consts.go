package mongo

const (
	REGISTRY_AUTH_DB          = "registry_auth"
	REG_AUTH_USER_COLL        = REGISTRY_AUTH_DB + ":user"
	REG_AUTH_PROJECT_COLL     = REGISTRY_AUTH_DB + ":project"
	REG_AUTH_PROJ_MEMBER_COLL = REGISTRY_AUTH_DB + ":project_member"
	REG_AUTH_ROLE_COLL        = REGISTRY_AUTH_DB + ":role"
	REG_AUTH_REPO_COLL        = REGISTRY_AUTH_DB + ":repository"
	REG_AUTH_ACCESSLOG_COLL   = REGISTRY_AUTH_DB + ":access_log"
)
