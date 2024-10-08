package easyenv

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type EasyEnv struct {
	connections       []*Connection
	currentConnection *Connection
}

type Connection struct {
	Name      string               // db file name
	dbPath    string               // db absolute path (acts like an id too)
	db        *sql.DB              // db instance
	projects  map[string]*Project  // projects and the associated env data
	templates map[string]*Template // templates of all the envs
}

func NewEasyEnv() *EasyEnv {
	return new(EasyEnv)
}

/*
	Unexported methods
*/

func (easy *EasyEnv) getConnectionBydbPath(dbPath string) (*Connection, error) {
	for _, connection := range easy.connections {
		if connection.dbPath == dbPath {
			return connection, nil
		}
	}
	return nil, fmt.Errorf("no connection found for the database with the name: %s", dbPath)
}

func (easy *EasyEnv) removeConnection(dbPath string) {
	tmp := make([]*Connection, 0)
	foundIndex := 0
	for index, connection := range easy.connections {
		if connection.dbPath == dbPath {
			foundIndex = index
			break
		}
	}
	tmp = append(tmp, easy.connections[:foundIndex]...)
	tmp = append(tmp, easy.connections[foundIndex+1:]...)
	easy.connections = tmp
}

func (easy *EasyEnv) isCurrentDBSet() error {

	if easy.currentConnection == nil {
		return fmt.Errorf("no database is currently open. Please open a database first using 'Open(path/to/sqlitefile)' before making any other calls")
	}

	return nil
}

func (easy *EasyEnv) Load(dbPath string) (*Connection, error) {
	db, err := sql.Open("sqlite3", dbPath)

	connection := new(Connection)

	if err != nil {
		return nil, err
	}

	connection.dbPath = dbPath
	connection.db = db
	splittedPath := strings.Split(dbPath, string(os.PathSeparator))
	connection.Name = splittedPath[len(splittedPath)-1]
	connection.projects = make(map[string]*Project)
	connection.templates = make(map[string]*Template)

	easy.connections = append(easy.connections, connection)
	easy.currentConnection = connection
	return easy.currentConnection, nil
}

func (easy *EasyEnv) Open(dbPath string) (*Connection, error) {
	connection, err := easy.getConnectionBydbPath(dbPath)

	if err != nil {
		return nil, err
	}

	easy.currentConnection = connection
	return connection, nil
}

func (easy *EasyEnv) CloseDB(dbPath string) error {
	connection, err := easy.getConnectionBydbPath(dbPath)

	if err != nil {
		return err
	}

	err = connection.db.Close()

	if err != nil {
		return err
	}

	easy.removeConnection(dbPath)

	if easy.currentConnection.dbPath == dbPath {
		easy.currentConnection = nil
	}

	return nil
}

func (easy *EasyEnv) CreateNewDB(dbPath string) (*Connection, error) {
	connection, err := easy.Load(dbPath)

	if err != nil {
		return nil, err
	}

	err = createTables(connection)

	if err != nil {
		return nil, err
	}

	return connection, nil
}

func (easy *EasyEnv) SaveDB() error {

	err := easy.isCurrentDBSet()

	if err != nil {
		return err
	}

	err = saveDataInDB(easy.currentConnection)

	if err != nil {
		return err
	}

	err = easy.SaveAllProjectEnvironmentsToFile()

	if err != nil {
		return err
	}

	easy.currentConnection.projects, err = easy.LoadProjects()

	if err != nil {
		return err
	}

	easy.currentConnection.templates, err = easy.LoadTemplates()

	if err != nil {
		return err
	}

	return nil
}

func (easy *EasyEnv) SaveAllProjectEnvironmentsToFile() error {

	err := easy.isCurrentDBSet()

	if err != nil {
		return err
	}

	for _, project := range easy.currentConnection.projects {
		err = project.SaveEnvironmentsToFile()

		if err != nil {
			return err
		}
	}

	return nil
}

func (easy *EasyEnv) AddProject(projectName, path string) (*Project, error) {
	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	project := NewProject(projectName, path)

	easy.currentConnection.projects[project.projectID] = project

	return project, nil
}

func (easy *EasyEnv) AddTemplate(templateName string) (*Template, error) {
	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	template := NewTemplate(templateName)

	easy.currentConnection.templates[template.templateID] = template

	return template, nil
}

/*
 Getters
*/

func (easy *EasyEnv) LoadProjects() (map[string]*Project, error) {
	projects, err := selectProjects(easy.currentConnection)

	if err != nil {
		return nil, err
	}

	easy.currentConnection.projects = projects

	for _, project := range projects {

		err := project.LoadEnvironmentsFromFile()

		if err != nil {
			return projects, err
		}
	}

	return projects, nil
}

func (easy *EasyEnv) LoadTemplates() (map[string]*Template, error) {
	templates, err := selectTemplates(easy.currentConnection)

	if err != nil {
		return nil, err
	}

	easy.currentConnection.templates = templates

	return templates, nil
}

func (easy *EasyEnv) AddTemplateEnvsToProject(templateID, projectID string) error {
	project, err := easy.GetProject(projectID)

	if err != nil {
		return err
	}

	template, err := easy.GetTemplate(templateID)

	if err != nil {
		return err
	}

	envs := template.GetEnvironments()

	for _, env := range envs {
		project.AddEnvironment(env.GetKey(), env.GetValue())
	}

	return nil
}

func (easy *EasyEnv) GetDatabases() []*Connection {
	return easy.connections
}

func (easy *EasyEnv) GetProject(projectID string) (*Project, error) {

	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	project, ok := easy.currentConnection.projects[projectID]

	if !ok {
		return nil, fmt.Errorf("no project found with ID %s. Please check the ID and try again", projectID)
	}

	return project, nil
}

func (easy *EasyEnv) GetProjects() (map[string]*Project, error) {
	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	return easy.currentConnection.projects, nil
}

func (easy *EasyEnv) GetTemplate(templateID string) (*Template, error) {
	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	template, ok := easy.currentConnection.templates[templateID]

	if !ok {
		return nil, fmt.Errorf("no template found with ID %s. Please verify the ID and try again", templateID)
	}

	return template, nil
}

func (easy *EasyEnv) GetTemplates() (map[string]*Template, error) {
	err := easy.isCurrentDBSet()

	if err != nil {
		return nil, err
	}

	return easy.currentConnection.templates, nil
}
