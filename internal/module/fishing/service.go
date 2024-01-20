package fishing

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type fishingService struct {
	db     IAquariumDB
	rng    *rand.Rand
	logger *zap.Logger

	creatures         []*Creature
	rarities          []*CreatureRarity
	weightedRaritySum int
}

type creatureWithCaption struct {
	*Creature
	caption string
}

func newFishingService(db IAquariumDB, logger *zap.Logger) (*fishingService, error) {
	s := &fishingService{
		db:     db,
		rng:    rand.New(rand.NewSource(time.Now().Unix())),
		logger: logger.Named("service"),
	}

	var err error
	if s.creatures, err = s.db.GetCreatures(); err != nil {
		return nil, err
	}
	// if really wanted, rarities can be extracted from creatures instead.
	if s.rarities, err = s.db.GetCreatureRarities(); err != nil {
		return nil, err
	}
	for _, r := range s.rarities {
		s.weightedRaritySum += r.Weight
	}

	return s, nil
}

func (fs *fishingService) getOrCreateAquarium(userID string) (*Aquarium, error) {
	aq, err := fs.db.GetAquarium(userID)
	if err != nil && err == sql.ErrNoRows {
		if err = fs.db.CreateAquarium(userID); err == nil {
			aq, err = fs.db.GetAquarium(userID)
		}
	}
	if err != nil {
		fs.logger.Error("could not get or create aquarium", zap.Error(err))
	}
	return aq, err
}

func (fs *fishingService) goFishing(userID string) (*creatureWithCaption, error) {
	aq, err := fs.getOrCreateAquarium(userID)
	if err != nil {
		return nil, err
	}

	r := fs.getRandomRarity()
	c := fs.getRandomCreature(r)

	caption := ""
	switch c.Name {
	case "Common":
		aq.Common++
		caption = fmt.Sprintf("You got a common - %v", c.DisplayName)
	case "Uncommon":
		aq.Uncommon++
		caption = fmt.Sprintf("You got an uncommon - %v", c.DisplayName)
	case "Rare":
		aq.Rare++
		caption = fmt.Sprintf("Ohhh, you got a rare! - %v", c.DisplayName)
	case "Super rare":
		aq.SuperRare++
		caption = fmt.Sprintf("Woah! you got a super rare! - %v", c.DisplayName)
	case "Legendary":
		aq.Legendary++
		caption = fmt.Sprintf("No way, you got a LEGENDARY!! - %v", c.DisplayName)
	}

	if err = fs.db.UpdateAquarium(aq); err != nil {
		return nil, err
	}
	return &creatureWithCaption{c, caption}, nil
}

func (fs *fishingService) getRandomRarity() *CreatureRarity {
	pick := fs.rng.Intn(fs.weightedRaritySum)
	for _, r := range fs.rarities {
		if pick < r.Weight {
			return r
		}
		pick -= r.Weight
	}
	return nil
}

func (fs *fishingService) getRandomCreature(r *CreatureRarity) *Creature {
	// FIXME: consider calculating this once in newFishingService() instead.
	var cs []*Creature
	for _, c := range fs.creatures {
		if c.Rarity.UID == r.UID {
			cs = append(cs, c)
		}
	}
	pick := fs.rng.Intn(len(cs))
	return cs[pick]
}
