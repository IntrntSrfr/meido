package fishing

type Aquarium struct {
	UserID    string `db:"user_id"`
	Common    int    `db:"common"`
	Uncommon  int    `db:"uncommon"`
	Rare      int    `db:"rare"`
	SuperRare int    `db:"super_rare"`
	Legendary int    `db:"legendary"`
}

type CreatureRarity struct {
	UID         int    `db:"uid"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Weight      int    `db:"weight"`
}

type Creature struct {
	UID         int             `db:"uid"`
	Name        string          `db:"name"`
	DisplayName string          `db:"display_name"`
	RarityID    int             `db:"rarity_id"`
	Rarity      *CreatureRarity `db:"-"`
}
