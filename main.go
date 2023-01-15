package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123456"
	dbname   = "postgres"
)

var db *sql.DB
var maxUserID int = 1
var maxProjectID int = 1
var maxTaskID int = 4

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetBody(r *http.Request) []byte {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	CheckError(err)
	return b
}

type Users struct {
	UsersList []string `json:"users"`
}

type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SuccussfulAuthData struct {
	Status bool `json:"status"`
	IsRoot bool `json:"isRoot"`
}

type UnsuccessfuleAuthData struct {
	Status       bool   `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

type UserProjectData struct {
	ProjectID int    `json:"projectId"`
	UserLogin string `json:"userLogin"`
}

type ProjectData struct { // ne ok
	ID           int    `json:"id"`
	Manager      string `json:"manager"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsArchive    bool   `json:"isArchive"`
	CreationDate string `json:"createdDate"`
}

type ProjectDataList struct {
	Projects []ProjectData `json:"projects"`
}

type ProjectIDData struct {
	ProjectID int `json:"projectId"`
}

type UserLoginData struct {
	UserLogin string `json:"userLogin"`
}

type UserRoles struct {
	Roles []string `json:"roles"`
}

type Task struct {
	ID       int    `json:"id"`
	Author   string `json:"author"`
	Assignee string `json:"assignee"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

type Attachment struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type CreateTaskInfo struct {
	IsTesting   bool     `json:"isTesting"`
	ProjectID   int      `json:"projectId"`
	Author      string   `json:"author"`
	Asignee     string   `json:"asignee"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	Attachments []string `json:"attachments"`
}

type ChangeTaskInfo struct {
	TaskID         int          `json:"taskId"`
	ProjectID      int          `json:"projectId"`
	Asignee        string       `json:"asignee"`
	Name           string       `json:"name"`
	Status         string       `json:"status"`
	Description    string       `json:"description"`
	AttachmentsOld []Attachment `json:"attachmentsOld"`
	AttachmentsNew []string     `json:"attachmentsNew"`
}

type ChangeRoleInfo struct {
	ProjectID int      `json:"projectId"`
	UserLogin string   `json:"userLogin"`
	Roles     []string `json:"roles"`
}

type TaskInfo struct {
	ID                   int          `json:"id"`
	Type                 string       `json:"type"`
	Author               string       `json:"author"`
	Asignee              string       `json:"asignee"`
	Name                 string       `json:"name"`
	Status               string       `json:"status"`
	Description          string       `json:"description"`
	AvailableTransitions []string     `json:"availableTransitions"`
	Attachments          []Attachment `json:"attachments"`
}

type ResultData struct {
	Status bool `json:"status"`
}

type ProjectTaskData struct {
	ProjectID int `json:"projectId"`
	TaskID    int `json:"taskId"`
}

type SuitesList struct {
	Suits []string `json:"suits"`
}

type TestCasesList struct {
	Cases []string `json:"cases"`
}

type TestRunsList struct {
	Runs []string `json:"runs"`
}

type LinkData struct {
	Link string `json:"link"`
}

type TaskTransitData struct {
	ProjectID int    `json:"projectId"`
	TaskID    int    `json:"taskId"`
	Status    string `json:"status"`
}

func initMaxUserID() {
	query := `select max(id) from users`
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		err = rows.Scan(&maxUserID)
		maxUserID++
		return
	}
}

func initMaxProjectID() {
	query := `select max(id) from projects`
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		err = rows.Scan(&maxProjectID)
		maxProjectID++
		return
	}
}

func initMaxTaskID() {
	query := `select max(task_id) from tasks`
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		err = rows.Scan(&maxTaskID)
		maxTaskID++
		return
	}
}

func preprocessRequest(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
		(*w).Header().Set("Access-Control-Allow-Headers", "X-Requested-With,content-type")
		(*w).Header().Set("Access-Control-Allow-Credentials", "true")
		fmt.Fprintf(*w, "")
	}
}

func getUserIDByLogin(login string) int {
	query := fmt.Sprintf(`select id from users where users.login = '%s'`, login)
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		var id int

		err = rows.Scan(&id)
		CheckError(err)
		return id
	}
	return -1
}

func isUserRoot(userID int) bool {
	query := fmt.Sprintf(`select is_root from users where id = %d`, userID)
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		var isRoot bool

		err = rows.Scan(&isRoot)
		return isRoot
	}
	return false
}

func sayHello(w http.ResponseWriter, r *http.Request) {

}

func getUsers(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	query := `SELECT "login" FROM "users"`
	rows, err := db.Query(query)
	CheckError(err)

	var response Users
	for rows.Next() {
		var login string

		err = rows.Scan(&login)
		CheckError(err)

		response.UsersList = append(response.UsersList, login)
	}
	jsonResp, err := json.Marshal(response)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var loginData LoginData
	err := json.Unmarshal(body, &loginData)
	CheckError(err)

	query := fmt.Sprintf(`select password, is_root from users where login = '%s'`, loginData.Username)
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		var password string
		var isRoot bool

		err = rows.Scan(&password, &isRoot)
		CheckError(err)
		var jsonResp []byte
		if password == loginData.Password {
			var response = &SuccussfulAuthData{
				Status: true,
				IsRoot: isRoot,
			}
			jsonResp, err = json.Marshal(*response)
			CheckError(err)
		} else {
			var response = &UnsuccessfuleAuthData{
				Status:       false,
				ErrorMessage: "Неправильный пароль",
			}
			jsonResp, err = json.Marshal(*response)
			CheckError(err)
		}
		fmt.Fprintf(w, string(jsonResp))
		break
	}
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userData UserData
	err := json.Unmarshal(body, &userData)
	CheckError(err)

	query := `INSERT INTO "users" ("id", "login", "password", "name", "is_root", "email") VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.Exec(query, maxUserID, userData.Login, userData.Password, "test", false, "test")
	CheckError(err)
	maxUserID++
	w.WriteHeader(201)
	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func attachUser(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userProjectData UserProjectData
	err := json.Unmarshal(body, &userProjectData)
	CheckError(err)

	userId := getUserIDByLogin(userProjectData.UserLogin)

	query := `INSERT INTO "project_users" ("user_id", "project_id", "role_name") VALUES ($1, $2, $3)`
	_, err = db.Exec(query, userId, userProjectData.ProjectID, "Пользователь")
	CheckError(err)

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func detachUser(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userProjectData UserProjectData
	err := json.Unmarshal(body, &userProjectData)
	CheckError(err)

	userId := getUserIDByLogin(userProjectData.UserLogin)

	query := `delete from "project_users" where user_id = $1 and project_id = $2`
	_, err = db.Exec(query, userId, userProjectData.ProjectID)
	CheckError(err)

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func createProject(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectData ProjectData
	err := json.Unmarshal(body, &projectData)
	CheckError(err)

	managerID := getUserIDByLogin(projectData.Manager)

	query := `insert into projects (id, manager_id, name, description, is_archive, creation_date) values($1, $2, $3, $4, $5, $6)`
	_, err = db.Exec(query, maxProjectID, managerID, projectData.Name, projectData.Description, projectData.IsArchive, "2022-01-16")
	CheckError(err)

	// add user to the project with manager and user roles

	query = `INSERT INTO "project_users" ("user_id", "project_id", "role_name") VALUES ($1, $2, $3)`
	_, err = db.Exec(query, managerID, maxProjectID, "Руководитель проекта")
	CheckError(err)
	_, err = db.Exec(query, managerID, maxProjectID, "Пользователь")
	CheckError(err)

	maxProjectID++

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	CheckError(err)
	w.WriteHeader(201)
	fmt.Fprintf(w, string(jsonResp))
}

func changeProject(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectData ProjectData
	err := json.Unmarshal(body, &projectData)
	CheckError(err)

	managerID := getUserIDByLogin(projectData.Manager)

	query := `update projects set manager_id = $1, name = $2, description = $3, is_archive = $4
	where id = $5`
	_, err = db.Query(query, managerID, projectData.Name, projectData.Description, projectData.IsArchive, projectData.ID)
	CheckError(err)

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func getProjectUsers(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectIDData ProjectIDData
	err := json.Unmarshal(body, &projectIDData)
	CheckError(err)

	query := fmt.Sprintf(`select distinct users.login from project_users 
	inner join users on users.id = project_users.user_id
	where project_users.project_id = '%d'`, projectIDData.ProjectID)
	rows, err := db.Query(query)
	CheckError(err)

	var users Users
	for rows.Next() {
		var login string

		rows.Scan(&login)
		users.UsersList = append(users.UsersList, login)
	}
	jsonResp, err := json.Marshal(users)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getUserProjectRoles(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userProjectData UserProjectData
	err := json.Unmarshal(body, &userProjectData)
	CheckError(err)

	userID := getUserIDByLogin(userProjectData.UserLogin)
	query := fmt.Sprintf(`select role_name from project_users where user_id = '%d' and project_id = '%d'`, userID, userProjectData.ProjectID)
	rows, err := db.Query(query)
	CheckError(err)

	var userRoles UserRoles
	for rows.Next() {
		var role string
		rows.Scan(&role)
		userRoles.Roles = append(userRoles.Roles, role)
	}
	jsonResp, err := json.Marshal(userRoles)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getUserProjects(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userLoginData UserLoginData
	err := json.Unmarshal(body, &userLoginData)
	CheckError(err)

	userID := getUserIDByLogin(userLoginData.UserLogin)

	var query string
	if isUserRoot(userID) {
		query = `select distinct projects.id, users.login, projects.name,
		projects.description, creation_date, is_archive 
		from projects
		inner join users on projects.manager_id = users.id`
	} else {
		query = fmt.Sprintf(`select distinct projects.id, users.login, projects.name,
		projects.description, creation_date, is_archive 
		from project_users 
		inner join projects on project_users.project_id = projects.id 
		inner join users on projects.manager_id = users.id
		where project_users.user_id = %d`, userID)
	}

	rows, err := db.Query(query)
	CheckError(err)

	var dataList ProjectDataList
	for rows.Next() {
		var id int
		var manager string
		var name string
		var description string
		var createdDate string
		var isArchive bool

		rows.Scan(&id, &manager, &name, &description, &createdDate, &isArchive)

		project := &ProjectData{
			ID:           id,
			Manager:      manager,
			Name:         name,
			Description:  description,
			IsArchive:    isArchive,
			CreationDate: createdDate,
		}
		dataList.Projects = append(dataList.Projects, *project)
	}
	if len(dataList.Projects) == 0 {
		dataList.Projects = make([]ProjectData, 0, 1)
	}
	jsonResp, err := json.Marshal(dataList)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getProjectTasks(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectIDData ProjectIDData
	err := json.Unmarshal(body, &projectIDData)
	CheckError(err)

	projectID := projectIDData.ProjectID

	query := fmt.Sprintf(`select task_id, users1.login, users2.login, tasks.name, status_name from tasks 
		inner join users as users1 on author_id = users1.id
		inner join users as users2 on asignee_id = users2.id
		where project_id = %d`, projectID)
	rows, err := db.Query(query)
	CheckError(err)

	var taskList TaskList
	for rows.Next() {
		var taskID int
		var creatorLogin string
		var asigneeLogin string
		var name string
		var status string

		err = rows.Scan(&taskID, &creatorLogin, &asigneeLogin, &name, &status)
		CheckError(err)

		task := &Task{
			ID:       taskID,
			Author:   creatorLogin,
			Assignee: asigneeLogin,
			Name:     name,
			Status:   status,
		}
		taskList.Tasks = append(taskList.Tasks, *task)
	}
	jsonResp, err := json.Marshal(taskList)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getProjectTesters(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectIDData ProjectIDData
	err := json.Unmarshal(body, &projectIDData)
	CheckError(err)

	query := fmt.Sprintf(`select users.login from project_users
	inner join users on project_users.user_id = users.id
	where project_users.role_name = 'Тестировщик' and project_users.project_id = %d`, projectIDData.ProjectID)
	rows, err := db.Query(query)
	CheckError(err)

	var response Users
	for rows.Next() {
		var login string

		err = rows.Scan(&login)
		CheckError(err)

		response.UsersList = append(response.UsersList, login)
	}
	jsonResp, err := json.Marshal(response)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getTask(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)
	body := GetBody(r)
	var projectTaskData ProjectTaskData
	err := json.Unmarshal(body, &projectTaskData)
	CheckError(err)

	query := fmt.Sprintf(`select task_id, users1.login, users2.login, status_name, tasks.name, description
	from tasks
	inner join users as users1 on author_id = users1.id
	inner join users as users2 on asignee_id = users2.id
	where tasks.project_id = %d and tasks.task_id = %d`, projectTaskData.ProjectID, projectTaskData.TaskID)
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		var id int
		var author string
		var asignee string
		var status string
		var name string
		var description string

		err = rows.Scan(&id, &author, &asignee, &status, &name, &description)
		CheckError(err)

		taskInfo := &TaskInfo{
			ID:          id,
			Type:        "i don't know",
			Author:      author,
			Asignee:     asignee,
			Name:        name,
			Status:      status,
			Description: description,
		}

		// get next transitions

		query = fmt.Sprintf(`select next from transitions where previous = '%s'`, status)
		rows2, err := db.Query(query)
		CheckError(err)

		var nextTransitions []string
		for rows2.Next() {
			var nextStatus string
			err = rows2.Scan(&nextStatus)
			CheckError(err)

			nextTransitions = append(nextTransitions, nextStatus)
		}
		taskInfo.AvailableTransitions = nextTransitions

		// get file attachments
		query = fmt.Sprintf(`select file_path from attachments where project_id = %d and task_id = %d`, projectTaskData.ProjectID, projectTaskData.TaskID)
		rows2, err = db.Query(query)
		CheckError(err)

		var attachments []Attachment = make([]Attachment, 0, 1)
		for rows2.Next() {
			var filename string
			err = rows2.Scan(&filename)
			CheckError(err)

			attachments = append(attachments, Attachment{filename, "www.link.com"})
		}
		taskInfo.Attachments = attachments

		jsonResp, err := json.Marshal(taskInfo)
		CheckError(err)
		fmt.Fprint(w, string(jsonResp))
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var createTaskInfo CreateTaskInfo
	err := json.Unmarshal(body, &createTaskInfo)
	CheckError(err)

	authorID := getUserIDByLogin(createTaskInfo.Author)
	asigneeID := getUserIDByLogin(createTaskInfo.Asignee)
	if createTaskInfo.IsTesting {
		createTaskInfo.Description += " www.testme.com"
	}

	// create task

	query := `insert into tasks (project_id, task_id, author_id, asignee_id, status_name, name, description)
	values ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.Exec(query, createTaskInfo.ProjectID, maxTaskID, authorID, asigneeID, "Новая задача", createTaskInfo.Name, createTaskInfo.Description)
	CheckError(err)

	// create attachments
	query = `insert into attachments (project_id, task_id, file_path, attachment_date) values ($1, $2, $3, $4)`
	for _, attachment := range createTaskInfo.Attachments {
		_, err = db.Exec(query, createTaskInfo.ProjectID, maxTaskID, attachment, "2022-01-16")
		CheckError(err)
	}

	maxTaskID++

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func hasUserSuchRole(userID int, projectID int, role string) bool {
	query := fmt.Sprintf(`select * from project_users where project_id = %d and user_id = %d and role_name = '%s'`,
		projectID, userID, role)
	rows, err := db.Query(query)
	CheckError(err)
	for rows.Next() {
		return true
	}
	return false
}

func changeUserRoles(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)
	body := GetBody(r)
	var changeRoleInfo ChangeRoleInfo
	err := json.Unmarshal(body, &changeRoleInfo)
	CheckError(err)

	userID := getUserIDByLogin(changeRoleInfo.UserLogin)

	// delete all user role in the project
	query := `delete from project_users where user_id = $1 and project_id = $2`
	_, err = db.Exec(query, userID, changeRoleInfo.ProjectID)
	CheckError(err)

	for _, role := range changeRoleInfo.Roles {
		if hasUserSuchRole(userID, changeRoleInfo.ProjectID, role) {
			continue
		}

		query = `insert into project_users(user_id, project_id, role_name) values($1, $2, $3)`
		_, err = db.Exec(query, userID, changeRoleInfo.ProjectID, role)
		CheckError(err)
	}

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func changeTask(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var changeTaskInfo ChangeTaskInfo
	err := json.Unmarshal(body, &changeTaskInfo)
	CheckError(err)

	asigneeID := getUserIDByLogin(changeTaskInfo.Asignee)

	query := `update tasks set project_id = $1, asignee_id = $2, name = $3, description = $4, status_name = $5 
	where task_id = $6`
	_, err = db.Exec(query, changeTaskInfo.ProjectID, asigneeID, changeTaskInfo.Name,
		changeTaskInfo.Description, changeTaskInfo.Status, changeTaskInfo.TaskID)
	CheckError(err)

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	fmt.Fprintf(w, string(jsonResp))
}

func getSuitesList(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	suitesList := &SuitesList{
		Suits: []string{"suit1", "suit2", "suit3"},
	}
	jsonResp, err := json.Marshal(*suitesList)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getTestCasesList(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)
	testCasesList := &TestCasesList{
		Cases: []string{"case1", "case2", "case3"},
	}
	jsonResp, err := json.Marshal(*testCasesList)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func getTestRunsList(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)
	testRunsList := &TestRunsList{
		Runs: []string{"run1", "run2", "run3"},
	}
	jsonResp, err := json.Marshal(*testRunsList)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func generateLinkForTesting(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)
	testURL := &LinkData{
		Link: "www.testme.com",
	}
	jsonResp, err := json.Marshal(*testURL)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func transitTask(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var taskTransitData TaskTransitData
	err := json.Unmarshal(body, &taskTransitData)
	CheckError(err)

	// we should check if we can change task status
	// get current status
	query := fmt.Sprintf(`select status_name from tasks where project_id = %d and task_id = %d`,
		taskTransitData.ProjectID, taskTransitData.TaskID)
	rows, err := db.Query(query)
	var currentStatus string
	for rows.Next() {
		err = rows.Scan(&currentStatus)
		CheckError(err)
	}

	// check if we can change status

	query = fmt.Sprintf(`select next from transitions where previous = '%s'`, currentStatus)
	rows, err = db.Query(query)
	CheckError(err)
	var count int
	for rows.Next() {
		var status string
		err = rows.Scan(&status)
		CheckError(err)
		if status == taskTransitData.Status {
			count++
			break
		}
	}
	if count == 0 {
		result := &ResultData{Status: false}
		jsonResp, err := json.Marshal(result)
		CheckError(err)
		fmt.Fprintf(w, string(jsonResp))
		return
	}

	// change task status

	query = `update tasks
	set status_name = $1
	where project_id = $2 and task_id = $3`
	_, err = db.Exec(query, taskTransitData.Status, taskTransitData.ProjectID, taskTransitData.TaskID)
	CheckError(err)

	result := &ResultData{Status: true}
	jsonResp, err := json.Marshal(result)
	CheckError(err)
	fmt.Fprintf(w, string(jsonResp))
}

func main() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	initMaxUserID()
	initMaxProjectID()
	initMaxTaskID()

	http.HandleFunc("/", sayHello)
	http.HandleFunc("/user/login", loginUser)                          // ok
	http.HandleFunc("/user/register", registerUser)                    // ok
	http.HandleFunc("/projects", getUserProjects)                      // ok
	http.HandleFunc("/projects/create", createProject)                 // ok
	http.HandleFunc("/projects/change", changeProject)                 // ok
	http.HandleFunc("/user/attach", attachUser)                        // ok
	http.HandleFunc("/user/detach", detachUser)                        // ok
	http.HandleFunc("/user/roles/get", getUserProjectRoles)            // ok
	http.HandleFunc("/user/roles/change", changeUserRoles)             // ok
	http.HandleFunc("/projects/users", getProjectUsers)                // ok
	http.HandleFunc("/users", getUsers)                                // ok
	http.HandleFunc("/tasks", getProjectTasks)                         // ok
	http.HandleFunc("/projects/testers", getProjectTesters)            // ok
	http.HandleFunc("/task", getTask)                                  // ok
	http.HandleFunc("/task/change", changeTask)                        // ok
	http.HandleFunc("/task/create", createTask)                        // ok
	http.HandleFunc("/projects/test/suites", getSuitesList)            // ok
	http.HandleFunc("/projects/test/cases", getTestCasesList)          // ok
	http.HandleFunc("/projects/test/runs", getTestRunsList)            // ok
	http.HandleFunc("/projects/test/generate", generateLinkForTesting) // ok
	http.HandleFunc("/task/transit", transitTask)                      // ok
	err = http.ListenAndServe(":8081", nil)                            // устанавливаем порт веб-сервера
	CheckError(err)
}
