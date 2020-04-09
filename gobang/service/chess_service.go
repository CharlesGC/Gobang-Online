package service

import (
	"fmt"
	"gobang/dto"
	"gobang/entity"
	"gobang/redis"
	"gobang/util"
	"log"
)

func SetReady(rid string, pid string, ready bool) (*entity.Room, error) {
	room, err := redis.GetRoom(rid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	inRoom, role, _ := isInRoom(pid, room)

	if !inRoom {
		err = fmt.Errorf("error: Player %v not in room %v", pid, rid)
		log.Println(err)
		return nil, err
	}

	if role == "host" {
		room.Host.Ready = ready
	} else if role == "challenger" {
		room.Challenger.Ready = ready
	} else {
		err = fmt.Errorf("error: Role %v cannot get ready", role)
		log.Println(err)
		return nil, err
	}
	room.Started = room.Host.Ready && room.Challenger.Ready
	if room.Started {
		room.Steps = make([]entity.Chess, 0)
	}

	if err = redis.SetRoom(room); err != nil {
		return nil, err
	}

	return room, nil
}

func MakeStep(rid string, c entity.Chess) (*entity.Room, error) {
	room, err := redis.GetRoom(rid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if room.Started {
		room.Steps = append(room.Steps, c)
	} else {
		err = fmt.Errorf("error: Can not make step while game is not started")
		log.Println(err)
		return nil, err
	}

	if err = redis.SetRoom(room); err != nil {
		log.Println(err)
		return nil, err
	}
	return room, nil
}

func PrepareNewGame(room *entity.Room) error {
	room.Host.Ready = false
	room.Challenger.Ready = false
	room.Host.Color = 1 - room.Host.Color
	room.Challenger.Color = 1 - room.Challenger.Color
	room.Started = false
	return redis.SetRoom(room)
}

func CheckFive(room *entity.Room) (bool, *dto.GameOverDTO) {
	hasFive, color := util.CheckFiveOfLastStep(&room.Steps)
	if !hasFive {
		return false, nil
	}

	var gameOverDTO *dto.GameOverDTO
	if room.Host.Color == color {
		gameOverDTO = &dto.GameOverDTO{
			RId:    room.Id,
			Winner: room.Host,
			Loser:  room.Challenger,
			Cause:  "five",
		}
	} else {
		gameOverDTO = &dto.GameOverDTO{
			RId:    room.Id,
			Winner: room.Challenger,
			Loser:  room.Host,
			Cause:  "five",
		}
	}

	PrepareNewGame(room)
	return true, gameOverDTO
}

func Surrender(pid string, rid string) (*dto.GameOverDTO, *entity.Room, error) {
	room, err := redis.GetRoom(rid)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	inRoom, role, _ := isInRoom(pid, room)
	if !inRoom || role == "spectator" {
		err = fmt.Errorf("error: player %v is not playing in room %v", pid, rid)
		log.Println(err)
		return nil, nil, err
	}

	gameOverDTO := &dto.GameOverDTO{
		RId:   rid,
		Cause: "surrender",
	}

	if role == "host" {
		gameOverDTO.Winner = room.Challenger
		gameOverDTO.Loser = room.Host
	} else if role == "challenger" {
		gameOverDTO.Winner = room.Host
		gameOverDTO.Loser = room.Challenger
	}

	if err = PrepareNewGame(room); err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return gameOverDTO, room, nil
}
