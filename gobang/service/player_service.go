package service

import (
	"gobang/entity"
	"gobang/redis"
	"log"
)

func NewPlayerConnect(id string) (*entity.Player, error) {
	p := &entity.Player{
		Id:     id,
		Name:   "unnamed",
		Status: "leisure",
	}
	err := redis.SetPlayer(p)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(id + " connects")
	return p, nil
}

func GetPlayer(id string) (*entity.Player, error) {
	return redis.GetPlayer(id)
}

func GetPlayers() (*[]entity.Player, error) {
	return redis.GetPlayers()
}

func PlayerDisconnect(id string) error {
	err := redis.DelPlayer(id)
	if err != nil {
		log.Println(err)
		return err
	}

	rooms, err := redis.GetRooms()
	if err != nil {
		log.Println(err)
		return err
	}

	for _, room := range *rooms {
		LeaveRoom(id, room.Id)
	}

	log.Println(id + " disconnects")
	return nil
}

func PlayerRename(id string, newName string) error {
	p, err := redis.GetPlayer(id)
	if err != nil {
		log.Println(err)
		return err
	}
	p.Name = newName
	err = redis.SetPlayer(p)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SetPlayerStatus(id string, status string) error {
	p, err := redis.GetPlayer(id)
	if err != nil {
		log.Println(err)
		return err
	}
	p.Status = status
	err = redis.SetPlayer(p)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
