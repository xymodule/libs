package db

import (
	"fmt"
	"math/rand"
	"testing"

	"db/db"
	pb "db/pb"
	"github.com/golang/protobuf/proto"
)

type User struct {
	Base pb.UserData
	DB   db.ProtoHash
}

func (p *User) GetId() int32 {
	return p.Base.GetId()
}

func (p *User) SetId(id int32) error {
	p.Base.Id = proto.Int32(id)
	return p.DB.SaveField("Id")
}

func (p *User) SetName(name string) error {
	p.Base.Name = proto.String(name)
	return p.DB.SaveField("Name")
}

func (p *User) GetName() string {
	return p.Base.GetName()
}

func (p *User) GetLevel() int32 {
	return p.Base.GetLevel()
}

func (p *User) SetLevel(level int32) error {
	p.Base.Level = proto.Int32(level)
	return p.DB.SaveField("Level")
}

func (p *User) AddItem(id int32) error {
	if p.Base.Items == nil {
		p.Base.Items = &pb.Ids{}
	}

	p.Base.Items.Id = append(p.Base.Items.Id, id)
	return p.DB.SaveField("Items")
}

func TestStorage(t *testing.T) {
	db.Init(false, true)

	user := &User{}
	user.DB.Bind(&user.Base, "User101")
	user.DB.Load()

	if user.GetId() == 0 {
		user.SetId(101)
		user.SetName("There")
		//user.Save()
	} else {
		fmt.Println(user)
		user.AddItem(int32(rand.Intn(100)))
	}

	/*
		client := db.GetRedis()

		checker := time.NewTicker(5 * time.Second)
		defer checker.Stop()

		for {
			select {
			case <-checker.C:
				gs, err := client.Get("test_redis01").Result()
				fmt.Println(gs, err)

				ss, err := client.Set("test_redis01", "hello", 0).Result()
				fmt.Println(ss, err)
			}
		}
	*/
}
