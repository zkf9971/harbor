use registry_auth;

db.access.insert({access_code: 'M', comment: 'Management access for project'});
db.access.insert({access_code: 'R', comment: 'Read access for project'});
db.access.insert({access_code: 'W', comment: 'Write access for project'});
db.access.insert({access_code: 'D', comment: 'Delete access for project'});
db.access.insert({access_code: 'S', comment: 'Search access for project'});

db.role.insert({role_id: 1, role_code: 'MDRWS', name: 'projectAdmin'});
db.role.insert({role_id: 2, role_code: 'RWS', name: 'developer'});
db.role.insert({role_id: 3, role_code: 'RS', name: 'guest'});


db.user.ensureIndex({
    "username": 1
},{
    "unique": true
})

db.user.ensureIndex({
    "email": 1
},{
    "unique": true
})


db.user.insert({
    username: 'admin',
    email: 'admin@example.com',
    password: '',
    realname: 'system admin',
    comment: 'admin user',
    deleted: 0,
    sysadmin_flag: 1,
    creation_time: new Date(),
    update_time: new Date()
});

db.user.insert({
    username: 'anonymous',
    email: 'anonymous@example.com',
    password: '',
    realname: 'anonymous user',
    comment: 'anonymous user',
    deleted: 1,
    sysadmin_flag: 0,
    creation_time: new Date(),
    update_time: new Date()
});


adminUser = db.user.findOne({username:'admin'});

db.project.ensureIndex({
    "name": 1
},{
    "unique": true
});


db.project.insert({
    owner_id: adminUser._id,
    name: 'library',
    creation_time: new Date(),
    update_time: new Date(),
    deleted: 0,
    public: 1
});

projectLibrary = db.project.findOne({name:'library'});


db.project_member.insert({
    project_id: projectLibrary._id,
    user_id: adminUser._id,
    role: 1,
    creation_time: new Date(),
    update_time: new Date()
});

db.access_log.ensureIndex({
    "project_id": 1,
    "op_time": 1
},
{
    "unique": true    
});

db.repository.ensureIndex({
    "name": 1
},
{
    "unique": true    
});




db.replication_job.ensureIndex({
    "policy_id": 1
},
{
    "unique": true    
});

db.replication_job.ensureIndex({
    "policy_id": 1,
    "update_time": 1
},
{
    "unique": true    
});



db.properties.ensureIndex({
    "k": 1
},
{
    "unique": true
});


db.alembic_version.insert({
    "version_num": "0.4.0"
});