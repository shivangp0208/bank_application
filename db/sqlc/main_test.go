package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shivangp0208/bank_application/util"
)

const (
	dbDriver = "mysql"
	dbSource = "root:12345678@tcp(127.0.0.1:3306)/bank_application?parseTime=true"
)

var store Store
var testDB *sql.DB
var err error

func init() {
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("unable to open db connection due to %v", err)
	}
}

func TestMain(m *testing.M) {
	store = NewStore(testDB)
	os.Exit(m.Run())
}

func GenerateRandomDBUser() (User, string) {
	pass := util.GenerateString(10)
	hashedPass, err := util.GenerateHashPassword(pass)
	if err != nil {
		panic(err)
	}
	return User{
		Username:          util.GenerateRandomUsername(10),
		HashedPassword:    hashedPass,
		Role:              util.User,
		FullName:          util.GenerateRandomFullName(6),
		Email:             util.GenerateRandomEmail(),
		IsVerified:        true,
		PasswordChangedAt: sql.NullTime{},
		CreatedAt:         time.Now(),
	}, pass
}

func GenerateRandomDBAccount(user User) Account {
	return Account{
		ID:        util.GenerateRandomID(),
		Owner:     user.Username,
		Balance:   util.GenerateRandomAmount(),
		Currency:  util.GenerateRandomCurrency(),
		CreatedAt: time.Now(),
	}
}
