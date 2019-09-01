package model

import "github.com/lmuench/gobdb/gobdb"

type Manager struct {
	ID int
	// Add attributes here
}

func (self *Manager) GetID() int {
	return self.ID
}

func (self *Manager) SetID(ID int) {
	self.ID = ID
}

func GetAllManagers(db *gobdb.DB) ([]Manager, error) {
	managers := []Manager{}

	err := db.GetAll(&Manager{}, func(resource interface{}) {
		manager := *resource.(*Manager)
		managers = append(managers, manager)
	})
	return managers, err
}

func GetAllManagersMap(db *gobdb.DB) (map[int]Manager, error) {
	managers := make(map[int]Manager)

	err := db.GetAll(&Manager{}, func(resource interface{}) {
		manager := *resource.(*Manager)
		managers[manager.GetID()] = manager
	})
	return managers, err
}
