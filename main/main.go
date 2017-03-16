package main

import (
	"fmt"
	"io/ioutil"

	"github.com/gregpechiro/csv"
)

func main() {

	/*	users := []User{
			{
				Name:     "name-1",
				Age:      1,
				Active:   true,
				Email:    "1@1.com",
				Password: "1",
				Address: Address{
					Street:  "1 - street",
					City:    "1 - city",
					County:  "1 - county",
					State:   "1 - state",
					Zip:     "11111",
					Country: "1 - country",
				},
			},
		}

		b, err := csv.Marshal(users)

		if err != nil {
			panic(err)
		}

		fmt.Printf("\n\n%s\n\n", b)

		var u []User
		if err := csv.Unmarshal(b, &u); err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n\n", u)*/

	b, err := ioutil.ReadFile("drivers.csv")
	if err != nil {
		panic(err)
	}

	var drivers []Driver
	if err := csv.Unmarshal(b, &drivers); err != nil {
		panic(err)
	}

	for _, driver := range drivers {
		fmt.Printf("%+v\n", driver)
	}

	b2, err := csv.Marshal(drivers)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("drivers2.csv", b2, 0666); err != nil {
		panic(err)
	}

	var driver Driver
	if err := csv.UnmarshalRow(5, b, &driver); err != nil {
		panic(err)
	}
	fmt.Printf("\n%+v\n", driver)
}

type User struct {
	Name     string  `csv:"name"`
	Age      int     `csv:"age"`
	Active   bool    `csv:"active"`
	Email    string  `csv:"email"`
	Password string  `csv:"-"`
	Address  Address `csv:"address"`
}

type Address struct {
	Street  string `csv:"street"`
	City    string `csv:"city"`
	County  string `csv:"-"`
	State   string `csv:"state"`
	Zip     string `csv:"zip"`
	Country string `csv:"-"`
}

type Driver struct {
	FirstName             string  `csv:"firstName"`
	LastName              string  `csv:"lastName"`
	DOB                   string  `csv:"DOB"`
	LicenseNumber         string  `csv:"licenseNum"`
	LicenseExpirationDate string  `csv:"licenseExpirationDate"`
	Last4OfSS             string  `csv:"last4OfSS"`
	DQF180                string  `csv:"DQF180"`
	DQF200                string  `csv:"DQF200"`
	Address               Address `csv:"address"`
	LicenseClass          string  `csv:"licenseclass"`
	CDLLicense            string  `csv:"cdlLicense"`
	MedicalExamExpires    string  `csv:"medicalExamExpires"`
	PullMVRBefore         string  `csv:"pullMVRBefore"`
	ProgramRenewalDate    string  `csv:"programRenewalDate"`
}
