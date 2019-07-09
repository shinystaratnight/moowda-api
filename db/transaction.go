package db

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

func Transaction(db *gorm.DB, txFunc func(*gorm.DB) error) (err error) {
	tx := db.Begin()

	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = errors.Wrap(p, "transaction error")
			default:
				err = errors.Errorf("transaction error: %s", p)
			}
		}

		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit().Error
	}()

	return txFunc(tx)
}
