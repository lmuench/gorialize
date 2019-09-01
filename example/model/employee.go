package model

import "github.com/lmuench/gobdb/gobdb"

type Employee struct {
	ID int
	ManagerID int
	// Add attributes here
}

func (self *Employee) GetID() int {
	return self.ID
}

func (self *Employee) SetID(ID int) {
	self.ID = ID
}

func (self Employee) GetManager(db *gobdb.DB) (Manager, error) {
	var manager Manager
	err := db.Get(&manager, self.ManagerID)
	return manager, err
}

func GetAllEmployees(db *gobdb.DB) ([]Employee, error) {
	employees := []Employee{}

	err := db.GetAll(&Employee{}, func(resource interface{}) {
		employee := *resource.(*Employee)
		employees = append(employees, employee)
	})
	return employees, err
}

func GetAllEmployeesMap(db *gobdb.DB) (map[int]Employee, error) {
	employees := make(map[int]Employee)

	err := db.GetAll(&Employee{}, func(resource interface{}) {
		employee := *resource.(*Employee)
		employees[employee.GetID()] = employee
	})
	return employees, err
}
