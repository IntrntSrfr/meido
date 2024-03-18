package feedback

import "github.com/intrntsrfr/meido/internal/database"

type IFeedbackDB interface {
	database.DB

	BlacklistUser(userID string) error
	UnblacklistUser(userID string) error
	IsUserBlacklisted(userID string) (bool, error)
}

type FeedbackDB struct {
	database.DB
}

func (db *FeedbackDB) BlacklistUser(userID string) error {
	_, err := db.Conn().Exec("INSERT INTO feedback_blacklist(user_id) VALUES($1);", userID)
	return err
}

func (db *FeedbackDB) UnblacklistUser(userID string) error {
	_, err := db.Conn().Exec("DELETE FROM feedback_blacklist WHERE user_id=$1;", userID)
	return err
}

func (db *FeedbackDB) IsUserBlacklisted(userID string) (bool, error) {
	var count int
	err := db.Conn().Get(&count, "SELECT COUNT(*) FROM feedback_blacklist WHERE user_id=$1;", userID)
	return count > 0, err
}

