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

// User holds the details of a user.
type User struct {
	UserID   bson.ObjectId `orm:"pk;column(user_id)" json:"user_id" bson:"_id,omitempty"`
	Username string        `orm:"column(username)" json:"username" bson:"username"`
	Email    string        `orm:"column(email)" json:"email" bson:"email"`
	Password string        `orm:"column(password)" json:"password" bson:"password"`
	Realname string        `orm:"column(realname)" json:"realname" bson:"realname"`
	Comment  string        `orm:"column(comment)" json:"comment" bson:"comment"`
	Deleted  int           `orm:"column(deleted)" json:"deleted" bson:"deleted"`
	Rolename string        `json:"role_name" bson:"-"`
	//if this field is named as "RoleID", beego orm can not map role_id
	//to it.
	Role int `json:"role_id"`
	//	RoleList     []Role `json:"role_list"`
	HasAdminRole int       `orm:"column(sysadmin_flag)" json:"has_admin_role" bson:"sysadmin_flag"`
	ResetUUID    string    `orm:"column(reset_uuid)" json:"reset_uuid" bson:"reset_uuid"`
	Salt         string    `orm:"column(salt)" json:"-" bson:"salt"`
	CreationTime time.Time `orm:"creation_time" json:"creation_time" bson:"creation_time"`
	UpdateTime   time.Time `orm:"update_time" json:"update_time" bson:"update_time"`
}
