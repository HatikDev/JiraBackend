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

// func sayHello(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello, Doroshin")
// }

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
	Tasks []Task `json:tasks`
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
	TaskID         int      `json:"taskId"`
	ProjectID      int      `json:"projectId"`
	Asignee        string   `json:"asignee"`
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Description    string   `json:"description"`
	AttachmentsOld []string `json:"attachmentsOld"`
	AttachmentsNew []string `json:"attachmentsNew"`
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

	query := fmt.Sprintf(`select password from users where login = '%s'`, loginData.Username)
	rows, err := db.Query(query)
	CheckError(err)

	for rows.Next() {
		var password string

		rows.Scan(&password)
		var jsonResp []byte
		if password == loginData.Password {
			var response = &SuccussfulAuthData{
				Status: true,
				IsRoot: false,
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
}

func attachUser(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var userProjectData UserProjectData
	err := json.Unmarshal(body, &userProjectData)
	CheckError(err)

	userId := getUserIDByLogin(userProjectData.UserLogin)

	query := `INSERT INTO "project_users" ("user_id", "project_id", "role_name") VALUES ($1, $2, $3)`
	_, err = db.Exec(query, userId, userProjectData.ProjectID, "Администратор")
	CheckError(err)
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
}

func createProject(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectData ProjectData
	err := json.Unmarshal(body, &projectData)
	CheckError(err)

	managerID := getUserIDByLogin(projectData.Manager)

	query := `insert into projects (id, manager_id, name, description, is_archive, creation_date) values($1, $2, $3, $4, $5, $6)`
	_, err = db.Exec(query, maxProjectID, managerID, projectData.Name, projectData.Description, projectData.IsArchive, "2017-04-03")
	CheckError(err)

	w.WriteHeader(201)
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
}

func getProjectUsers(w http.ResponseWriter, r *http.Request) {
	preprocessRequest(&w, r)

	body := GetBody(r)
	var projectIDData ProjectIDData
	err := json.Unmarshal(body, &projectIDData)
	CheckError(err)

	query := fmt.Sprintf(`select users.login from project_users 
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
	query := fmt.Sprintf(`select projects.id, users.login, projects.name, projects.description, creation_date, is_archive 
	from project_users 
	inner join projects on project_users.project_id = projects.id 
	inner join users on projects.manager_id = users.id
	where project_users.user_id = %d`, userID)
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

		query = fmt.Sprintf(`select next from transitions where previous = %s`, status)
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

		// w.Header().Set("Access-Control-Allow-Origin", "*")

		// get file attachments
		// query = fmt.Sprintf(`select file_path from attachments where project_id = %s and task_id = %s`, projectTaskData.ProjectID, projectTaskData.TaskID)
		// rows2, err = db.Query(query)
		// CheckError(err)

		// var attachments []string
		// for rows2.Next() {
		// 	var file string
		// 	err = rows2.Scan(&file)
		// 	CheckError(err)

		// 	attachments = append(attachments, file)
		// }
		// taskInfo.Attachments = attachments
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

	// create task

	query := `insert into tasks (project_id, task_id, author_id, asignee_id, status_name, name, description)
	values ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.Exec(query, createTaskInfo.ProjectID, maxTaskID, authorID, asigneeID, "Новая задача", createTaskInfo.Name, createTaskInfo.Description)
	CheckError(err)

	// create attachments
	query = `insert into attachments (project_id, task_id, file_path, attachment_date) values ($1, $2, $3, $4)`
	for _, attachment := range createTaskInfo.Attachments {
		_, err = db.Exec(query, createTaskInfo.ProjectID, maxTaskID, attachment, "2017-01-01")
		CheckError(err)
	}

	maxTaskID++
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

func main() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlconn)
	CheckError(err)

	defer db.Close()

	http.HandleFunc("/", sayHello)
	http.HandleFunc("/user/login", loginUser)               // ok
	http.HandleFunc("/user/register", registerUser)         // ok
	http.HandleFunc("/projects", getUserProjects)           // ok
	http.HandleFunc("/projects/create", createProject)      // ok
	http.HandleFunc("/projects/change", changeProject)      // ok
	http.HandleFunc("/user/attach", attachUser)             // ok
	http.HandleFunc("/user/detach", detachUser)             // ok
	http.HandleFunc("/user/roles/get", getUserProjectRoles) // ok
	http.HandleFunc("/user/roles/change", sayHello)
	http.HandleFunc("/projects/users", getProjectUsers)       // ok
	http.HandleFunc("/users", getUsers)                       // ok
	http.HandleFunc("/tasks", getProjectTasks)                // ok
	http.HandleFunc("/projects/testers", getProjectTesters)   // ok
	http.HandleFunc("/task", getTask)                         // not ok and we should fix attachments
	http.HandleFunc("/task/change", changeTask)               // what i should with it?
	http.HandleFunc("/task/create", createTask)               // ok
	http.HandleFunc("/projects/test/suites", getSuitesList)   // ok
	http.HandleFunc("/projects/test/cases", getTestCasesList) // ok
	http.HandleFunc("/projects/test/runs", getTestRunsList)   // ok
	http.HandleFunc("/projects/test/generate", sayHello)
	err = http.ListenAndServe(":8081", nil) // устанавливаем порт веб-сервера
	CheckError(err)
}
