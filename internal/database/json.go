package database

/*
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"os"
	"sync"
)

type JsonDB struct {
	path  string
	state *state
}

type state struct {
	sync.Mutex
	Guilds map[string]*Guild `json:"guilds"`
}

func NewJsonDatabase(path string) (*JsonDB, error) {
	db := &JsonDB{
		path: path,
		state: &state{
			Guilds: make(map[string]*Guild, 0),
		},
	}
	err := db.load(path)
	return db, err
}

func (j *JsonDB) Close() error {
	return j.save()
}

func (j *JsonDB) load(path string) error {
	if _, err := os.Stat(path); err != nil {
		// file does not exist, so use default
		fmt.Println("no data file found, using default")
		return nil
	}

	fmt.Println("data file found")
	d, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	state := &state{}
	err = json.Unmarshal(d, &state)
	if err != nil {
		return err
	}

	j.state = state
	return nil
}

func (j *JsonDB) save() error {
	d, err := json.Marshal(j.state)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(j.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(d)
	return err
}

func (j *JsonDB) GetConn() *sqlx.DB {
	return nil
}

func (j *JsonDB) CreateGuild(gid string) error {
	j.state.Lock()
	defer j.state.Unlock()
	if _, ok := j.state.Guilds[gid]; ok {
		return errors.New("key already exists")
	}
	g := &Guild{UID: gid}
	j.state.Guilds[gid] = g
	return nil
}

func (j *JsonDB) UpdateGuild(gid string, gc *Guild) error {
	j.state.Lock()
	defer j.state.Unlock()
	if _, ok := j.state.Guilds[gid]; !ok {
		return errors.New("key does not exist")
	}
	j.state.Guilds[gid] = gc
	return nil
}

func (j *JsonDB) GetGuild(gid string) (*Guild, error) {
	j.state.Lock()
	defer j.state.Unlock()
	if v, ok := j.state.Guilds[gid]; ok {
		return v, nil
	}
	return nil, errors.New("key does not exist")
}
*/
