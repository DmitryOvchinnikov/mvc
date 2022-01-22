package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Table base model of data represented in DB.
type Table struct {
	ID     uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	Text   string    `gorm:"column:text;type:varchar;size:255"`
	CodeID uuid.UUID `gorm:"column:code_id;type:uuid;not null"`
	Code   uint      `gorm:"column:code;type:smallint"`
	Name   string    `gorm:"column:name;type:varchar;size:255"`
}

type TableService interface {
	TableDB
}

type TableDB interface {
	ByID(id uuid.UUID) (*Table, error)
	ByCodeID(codeID uuid.UUID) ([]Table, error)
	Create(table *Table) error
	Update(table *Table) error
}

func NewTableService(db *gorm.DB) TableService {
	return &tableService{
		&tableValidator{&tableGORM{db}},
	}
}

type tableService struct {
	TableDB
}

type tableValidator struct {
	TableDB
}

func (t *tableValidator) Create(table *Table) error {
	err := runTableValFuncs(table,
		t.textRequired)
	if err != nil {
		return err
	}

	return t.TableDB.Create(table)
}

func (t *tableValidator) Update(table *Table) error {
	err := runTableValFuncs(table,
		t.textRequired)
	if err != nil {
		return err
	}

	return t.TableDB.Update(table)
}

func (t *tableValidator) textRequired(table *Table) error {
	if table.Text == "" {
		return ErrTextRequired
	}

	return nil
}

var _ TableDB = &tableGORM{}

type tableGORM struct {
	db *gorm.DB
}

func (t *tableGORM) ByID(id uuid.UUID) (*Table, error) {
	var table Table
	db := t.db.Where("id=?", id)
	err := first(db, &table)
	return &table, err
}

func (t *tableGORM) ByCodeID(codeID uuid.UUID) ([]Table, error) {
	var tables []Table
	err := t.db.Where("code_id=?", codeID).Find(&tables).Error
	if err != nil {
		return nil, err
	}
	return tables, err
}

func (t *tableGORM) Create(table *Table) error {
	return t.db.Create(table).Error
}

func (t *tableGORM) Update(table *Table) error {
	return t.db.Save(table).Error
}

// first will query using the provided gorm.DB and it will
// get the first item returned and place it into dst.
// If nothing is found in the query, it will return ErrNotFound.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

type tableValFunc func(*Table) error

func runTableValFuncs(table *Table, fns ...tableValFunc) error {
	for _, fn := range fns {
		err := fn(table)
		if err != nil {
			return err
		}
	}

	return nil
}
