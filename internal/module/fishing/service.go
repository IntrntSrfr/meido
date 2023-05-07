package fishing

import (
	"database/sql"
	"github.com/intrntsrfr/meido/internal/database"
	"math/rand"
)

type fishingService struct {
	db database.DB
}

type fishLevel int

const (
	common fishLevel = iota + 1
	uncommon
	rare
	superRare
	legendary
)

type Creature struct {
	level   fishLevel
	caption string
	mention bool
}

var creatures = []Creature{
	{common, "You got a common - 🐟", false},
	{uncommon, "You got an uncommon - 🐠", false},
	{rare, "Ohhh, you got a rare! - 🐡", false},
	{superRare, "Woah! you got a super rare! - 🦈", true},
	{legendary, "No way, you got a LEGENDARY!! - 🎷🦈", true},
}

func newFishingService(db database.DB) *fishingService {
	return &fishingService{db}
}

func (fs *fishingService) goFishing(userID string) (*Creature, error) {
	c := fs.getRandomCreature()
	aq, err := fs.db.GetAquarium(userID)
	if err != nil && err == sql.ErrNoRows {
		// if no aquarium found, make one
		err = fs.db.CreateAquarium(userID)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		// everything else we just return
		return nil, err
	}

	switch c.level {
	case common:
		aq.Common++
	case uncommon:
		aq.Uncommon++
	case rare:
		aq.Rare++
	case superRare:
		aq.SuperRare++
	case legendary:
		aq.Legendary++
	}

	if err = fs.db.UpdateAquarium(aq); err != nil {
		return nil, err
	}
	return c, nil
}

func (fs *fishingService) getRandomCreature() *Creature {
	pick := rand.Intn(1000) + 1
	var fp Creature
	if pick <= 800 {
		fp = creatures[0]
	} else if pick <= 940 {
		fp = creatures[1]
	} else if pick <= 990 {
		fp = creatures[2]
	} else if pick <= 999 {
		fp = creatures[3]
	} else {
		fp = creatures[4]
	}
	return &fp
}
