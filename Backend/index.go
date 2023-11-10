package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

func intializeMuxRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/signup", SignUp).Methods("POST")
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/addjobs", AddJobs).Methods("POST")
	r.HandleFunc("/getjobslist", GetJobsList)
	r.HandleFunc("/isAuthorized", isAuth)

	log.Fatal(http.ListenAndServe(":3000", r))

}
func main() {

	intializeMigration()

	intializeMuxRouter()

}
