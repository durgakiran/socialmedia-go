package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	file string
}

type databaseSchema struct {
	Users map[string]User `json:"users"`
	Posts map[string]Post `json:"posts"`
}

// User -
type User struct {
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
}

// Post -
type Post struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserEmail string    `json:"userEmail"`
	Text      string    `json:"text"`
}

func NewClient(filePath string) Client {
	fmt.Println("creating client" + filePath)
	newClient := Client{}
	newClient.file = filePath
	return newClient
}

func (c Client) createDb() error {
	data, err := json.Marshal(databaseSchema{})
	if err != nil {
		return err
	}
	err = os.WriteFile(c.file, data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) EnsureDb() error {
	_, err := os.ReadFile("testdata/hello")
	if err != nil {
		err = c.createDb()
	}

	return err
}

func (c Client) updateDB(db databaseSchema) error {
	data, _ := json.Marshal(db)
	err := os.WriteFile(c.file, data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (c Client) readDB() (databaseSchema, error) {
	data, err := os.ReadFile(c.file)
	if err != nil {
		return databaseSchema{}, err
	}
	var dbSchemaPointer databaseSchema
	err = json.Unmarshal(data, &dbSchemaPointer)
	if err != nil {
		return databaseSchema{}, err
	}
	return dbSchemaPointer, nil
}

func userIsEligible(email, password string, age int) error {
	if email == "" {
		return errors.New("email can't be empty")
	}

	if password == "" {
		return errors.New("password can't be empty")
	}

	if age < 18 {
		return errors.New("age must be at least 18 years old")
	}

	return nil
}

func (c Client) CreateUser(email, password, name string, age int) (User, error) {
	err := userIsEligible(email, password, age)

	if err != nil {
		return User{}, err
	}

	currentData, err := c.readDB()
	if err != nil {
		return User{}, err
	}

	user := User{
		Email:     email,
		Password:  password,
		Name:      name,
		Age:       age,
		CreatedAt: time.Now().UTC(),
	}

	// init if not initialised
	if currentData.Users == nil {
		currentData.Users = make(map[string]User)
	}

	_, ok := currentData.Users[email]
	if ok {
		return user, errors.New("User already exists")
	}

	currentData.Users[email] = user
	err = c.updateDB(currentData)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (c Client) UpdateUser(email, password, name string, age int) (User, error) {
	currentData, err := c.readDB()
	if err != nil {
		return User{}, err
	}
	user, ok := currentData.Users[email]
	if !ok {
		return User{}, errors.New("User doesn't exists")
	}

	user.Age = age
	user.Password = password
	user.Name = name

	currentData.Users[email] = user

	err = c.updateDB(currentData)
	if err != nil {
		return user, err
	}

	return user, nil

}

func (c Client) Getuser(email string) (User, error) {
	currentData, err := c.readDB()
	if err != nil {
		return User{}, err
	}
	user, ok := currentData.Users[email]
	if !ok {
		return User{}, errors.New("User doesn[t exists")
	}

	return user, nil
}

func (c Client) DeleteUser(email string) error {
	currentData, err := c.readDB()
	if err != nil {
		return err
	}

	delete(currentData.Users, email)

	err = c.updateDB(currentData)
	if err != nil {
		return err
	}

	return nil
}

func validatePostParams(userEmail string) error {
	if userEmail == "" {
		return errors.New("email can't be empty")
	}

	return nil
}

func (c Client) CreatePost(userEmail string, text string) (Post, error) {
	data, err := c.readDB()
	if err != nil {
		return Post{}, err
	}

	err = validatePostParams(userEmail)

	if err != nil {
		return Post{}, err
	}

	_, ok := data.Users[userEmail]

	if !ok {
		return Post{}, errors.New("User does not exist")
	}
	id := uuid.New().String()

	newPost := Post{
		ID:        id,
		CreatedAt: time.Now(),
		UserEmail: userEmail,
		Text:      text,
	}

	if data.Posts == nil {
		data.Posts = make(map[string]Post)
	}

	data.Posts[id] = newPost
	err = c.updateDB(data)
	if err != nil {
		return newPost, err
	}

	return newPost, nil
}

func (c Client) GetPosts(userEmail string) ([]Post, error) {
	data, err := c.readDB()
	if err != nil {
		return []Post{}, err
	}

	// init if not initialised
	if data.Posts == nil {
		data.Posts = make(map[string]Post)
	}

	posts := make([]Post, 0)

	for _, value := range data.Posts {
		if value.UserEmail == userEmail {
			posts = append(posts, value)
		}
	}

	return posts, nil

}

func (c Client) DeletePost(id string) error {
	currentData, err := c.readDB()
	if err != nil {
		return err
	}

	delete(currentData.Posts, id)

	err = c.updateDB(currentData)
	if err != nil {
		return err
	}

	return nil
}
