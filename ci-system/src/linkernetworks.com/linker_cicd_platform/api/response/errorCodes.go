package response

//error codes defined in xwiki
const (
	ErrCreateOpsenv      = "E20000"
	ErrQueryOpsenv       = "E20001"
	ErrUpdateOpsenv      = "E20002"
	ErrDeleteOpsenv      = "E20003"
	ErrCallController    = "E20004"
	ErrCreateProject     = "E20100"
	ErrDeleteProject     = "E20103"
	ErrProjectConflict   = "E20104"
	ErrProjectNotFound   = "E20105"
	ErrHideProject       = "E20106"
	ErrArtifactNotFound  = "E20204"
	ErrCreateJob         = "E20300"
	ErrBuildJob          = "E20305"
	
	ErrDockerFileExisted = "E20506"
	
	ErrBadRequestGeneric = "E20600"
	ErrBadRequestURL     = "E20601"
	ErrBadRequestBody    = "E20603"
	ErrUnauthorized      = "E20604"
	ErrConvertJson       = "E20703"
	ErrReadFile          = "E20705"
	ErrCreateFile        = "E20706"
	ErrDBInsert          = "E20713"
	ErrDBQuery           = "E20714"
	ErrDBDelete          = "E20715"
	ErrDBUpdate          = "E20716"
	ErrGenericInternal   = "E20700"
	ErrUpdateGerritPasswd = "E20801"
	ErrGerritAuthFailure  = "E20802"
	ErrUpdateLdapPasswd   = "E20803"
)

var errorMsg = map[string]string{
	ErrCreateOpsenv:      "Fail to create OpsEnv",
	ErrQueryOpsenv:       "Fail to query OpsEnv",
	ErrUpdateOpsenv:      "Fail to update OpsEnv",
	ErrDeleteOpsenv:      "Fail to delete OpsEnv",
	ErrCallController:    "Fail to call controller",
	ErrCreateProject:     "Fail to create project",
	ErrProjectConflict:   "Project already exists",
	ErrProjectNotFound:   "Project not found or not created via cicd platform",
	ErrDeleteProject:     "Fail to delete project",
	ErrHideProject:       "Fail to hide project",
	ErrArtifactNotFound:  "Artifact not found",
	ErrBadRequestURL:     "Bad request URL",
	ErrBadRequestBody:    "Bad request body",
	ErrUnauthorized:      "Fail to authorize",
	ErrConvertJson:       "Error marshal or unmarshal json",
	ErrDBInsert:          "Fail to insert into database",
	ErrDBQuery:           "Fail to query from database",
	ErrDBDelete:          "Fail to delete from database",
	ErrDBUpdate:          "Fail to update from database",
	ErrGenericInternal:   "Generic internal server error",
	ErrReadFile:          "Error read file",
	ErrCreateFile:        "Error create file",
	ErrBadRequestGeneric: "Bad request,generic",
	ErrCreateJob:         "Fail to create job in jenkins",
	ErrBuildJob:          "Fail to build job",
	ErrDockerFileExisted: "The docker file image is existed",
	ErrUpdateGerritPasswd: "Fail to update gerrit password",
	ErrGerritAuthFailure:  "Gerrit authorization failure, wrong username or password",
	ErrUpdateLdapPasswd:   "Fail to update openldap password",
}

//error codes to errormsg
func ErrorMsg(code string) string {
	return errorMsg[code]
}

//new Error
func NewError(code string) Error {
	if len(code) == 0 {
		return Error{}
	}
	return Error{code, ErrorMsg(code)}
}
