package fishing

import "github.com/intrntsrfr/meido/internal/database"

type IAquariumDB interface {
	database.DB

	CreateAquarium(userID string) error
	GetAquarium(userID string) (*Aquarium, error)
	UpdateAquarium(aquarium *Aquarium) error
	GetCreatureRarities() ([]*CreatureRarity, error)
	GetCreatures() ([]*Creature, error)
}

type AquariumDB struct {
	database.DB
}

func (db *AquariumDB) CreateAquarium(userID string) error {
	_, err := db.GetConn().Exec("INSERT INTO aquarium VALUES($1)", userID)
	return err
}

func (db *AquariumDB) GetAquarium(userID string) (*Aquarium, error) {
	var aquarium Aquarium
	err := db.GetConn().Get(&aquarium, "SELECT * FROM aquarium WHERE user_id=$1", userID)
	return &aquarium, err
}

func (db *AquariumDB) UpdateAquarium(aq *Aquarium) error {
	_, err := db.GetConn().Exec("UPDATE aquarium SET common=$1, uncommon=$2, rare=$3, super_rare=$4, legendary=$5 WHERE user_id=$6",
		aq.Common, aq.Uncommon, aq.Rare, aq.SuperRare, aq.Legendary, aq.UserID)
	return err
}

func (db *AquariumDB) GetCreatureRarities() ([]*CreatureRarity, error) {
	var rarities []*CreatureRarity
	err := db.GetConn().Select(&rarities, "SELECT * FROM creature_rarity")
	return rarities, err
}

func (db *AquariumDB) GetCreatures() ([]*Creature, error) {
	var creatures []*Creature
	err := db.GetConn().Select(&creatures, "SELECT * FROM creature")
	if err != nil {
		return nil, err
	}
	rarities, err := db.GetCreatureRarities()
	if err != nil {
		return nil, err
	}
	for _, c := range creatures {
		for _, r := range rarities {
			if c.RarityID == r.UID {
				c.Rarity = r
			}
		}
	}
	return creatures, nil
}
