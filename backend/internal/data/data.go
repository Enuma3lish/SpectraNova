package data
package data

import (
	"MLW/fenzVideo/internal/conf"
	"MLW/fenzVideo/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Data struct {
	db *gorm.DB
}

func NewData(c conf.Data, logger log.Logger) (*Data, error) {
	helper := log.NewHelper(logger)
	if c.Database.Driver != "mysql" {
		helper.Warnf("database driver %s not explicitly supported", c.Database.Driver)
	}










}	return &Data{db: db}, nil	helper.Info("database connected")	}		return nil, err	if err := db.AutoMigrate(&model.User{}); err != nil {	}		return nil, err	if err != nil {	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})