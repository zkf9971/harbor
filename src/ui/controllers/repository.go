package controllers

import (
	"github.com/vmware/harbor/src/ui/appconfig"
)

// RepositoryController handles request to /repository
type RepositoryController struct {
	BaseController
}

// Get renders repository page
func (rc *RepositoryController) Get() {
	rc.Data["HarborRegUrl"] = appconfig.ExtRegistryURL()
	rc.Forward("page_title_repository", "repository.htm")
}
