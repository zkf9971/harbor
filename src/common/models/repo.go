/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// RepoRecord holds the record of an repository in DB, all the infors are from the registry notification event.
type RepoRecord struct {
	RepositoryID bson.ObjectId `orm:"column(repository_id);pk" json:"repository_id" bson:"_id,omitempty"`
	Name         string        `orm:"column(name)" json:"name" bson:"name"`
	OwnerName    string        `orm:"-" bson:"-"`
	OwnerID      bson.ObjectId `orm:"column(owner_id)"  json:"owner_id" bson:"owner_id"`
	ProjectName  string        `orm:"-" bson:"-"`
	ProjectID    bson.ObjectId `orm:"column(project_id)"  json:"project_id" bson:"project_id"`
	Description  string        `orm:"column(description)" json:"description" bson:"description"`
	PullCount    int64         `orm:"column(pull_count)" json:"pull_count" bson:"pull_count"`
	StarCount    int64         `orm:"column(star_count)" json:"star_count" bson:"star_count"`
	CreationTime time.Time     `orm:"column(creation_time);auto_now_add" json:"creation_time" bson:"creation_time"`
	UpdateTime   time.Time     `orm:"column(update_time);auto_now" json:"update_time" bson:"update_time"`
}

//TableName is required by by beego orm to map RepoRecord to table repository
func (rp *RepoRecord) TableName() string {
	return "repository"
}
