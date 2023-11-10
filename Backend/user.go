package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type userSignUpDetails struct {
	gorm.Model
	UserName    string `gorm:"type:varchar(100);" json:"username"`
	Password    string `gorm:"type:varchar(255);" json:"password"`
	Role        string `gorm:"type:varchar(50);" json:"role"`
	Email       string `gorm:"type:varchar(255);" json:"email"`
	FirstName   string `gorm:"type:varchar(100);" json:"first_name"`
	LastName    string `gorm:"type:varchar(100);" json:"last_name"`
	PhoneNumber string `gorm:"type:varchar(20);" json:"phone_number"`
}

type loginDetails struct {
	UserName string `gorm:"type:varchar(100);" json:"username"`
	Password string `gorm:"type:varchar(255);" json:"password"`
}

type claims struct {
	UserName string `gorm:"type:varchar(100);" json:"username"`
	jwt.StandardClaims
}

type Response struct {
	ResponseCode int         `json:"response_code"`
	RespMessage  string      `json:"resp_message"`
	Data         interface{} `json:"data,omitempty"`
}

type ScanCredentials struct {
	UserName string
	Password string
	Role     string
}

type JobDetails struct {
	gorm.Model                // This includes fields ID, CreatedAt, UpdatedAt, DeletedAt
	JobTitle           string `json:"job_title" gorm:"column:job_title"`
	JobDescription     string `json:"job_description" gorm:"column:job_description"`
	ExperienceRequired int    `json:"experience_required" gorm:"column:experience_required"`
	CompanyName        string `json:"company_name" gorm:"column:company_name"`
	Location           string `json:"location" gorm:"column:location"`
	BondYears          int    `json:"bond_years" gorm:"column:bond_years"`
	PostedBy           uint   `json:"posted_by" gorm:"column:posted_by"`
}

type Data struct {
	JwtToken string `json :"jwt_token"`
	//JobsList             []JobDetails     `json:"jobs_list"`
	UserName string `json:"userName"`
	Role     string `json:"role"`
	//AppliedJoblist       []AppliedJobData `json:"applied_jobs"`
	//ApplieJobsByUserList []ApplyJob       `json:"user_jobs"`
	//JobDetailsList       []JobDetails     `json:"posted_jobs_list"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}

type ResponseGeneral struct {
	ResponseCode int
	RespMessage  string
}

type JobsListData struct {
	JobsList []JobDetails
}

var DB *gorm.DB
var err error
var jwtKey = []byte("secret_key")

// const DNS = "root:password123@tcp(127.0.0.1:3306)/golang_users" //charset not defined
const DNS = "root:password123@tcp(127.0.0.1:3306)/golang_users?charset=utf8mb4&parseTime=True"

func intializeMigration() {
	DB, err = gorm.Open(mysql.Open(DNS), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB.AutoMigrate(&userSignUpDetails{})
	DB.AutoMigrate(&JobDetails{})
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user userSignUpDetails
	json.NewDecoder(r.Body).Decode(&user)
	result := DB.Create(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logging in..")
	w.Header().Set("Content-Type", "application/json")
	var loginDet loginDetails
	err := json.NewDecoder(r.Body).Decode(&loginDet)
	fmt.Println(loginDet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var ExpectedCredentials ScanCredentials
	//var role string

	err = DB.Raw("SELECT user_name, password, role FROM user_sign_up_details WHERE user_name = ?", loginDet.UserName).
		Scan(&ExpectedCredentials).Error

	// If there's no user found in the database, return an error response
	if err == gorm.ErrRecordNotFound {
		http.Error(w, "User doesn't exist", http.StatusBadRequest)
		return
	}

	if ExpectedCredentials.Password != loginDet.Password {
		// If the passwords don't match, return an error response
		http.Error(w, "Please enter the correct password", http.StatusAccepted)
		return
	}
	fmt.Println("logged In successfully")
	expirationTime := time.Now().Add(time.Minute * 10)

	claims := &claims{
		UserName: loginDet.UserName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		// If there's an error generating the token, return an error response
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	var response Response
	response.ResponseCode = 200
	response.RespMessage = "User login Successful"
	//response.Data.JwtToken = tokenString

	response.Data = Data{

		Role:     ExpectedCredentials.Role,
		JwtToken: tokenString,
	}

	jsonResp, err := json.Marshal(response)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)

}
func AddJobs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Adding Job..")
	w.Header().Set("Content-Type", "application/json")

	var JobDet JobDetails

	err := json.NewDecoder(r.Body).Decode(&JobDet)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(Response{
			ResponseCode: http.StatusUnprocessableEntity,
			RespMessage:  "Invalid request payload",
		})
		return
	}
	result := DB.Create(&JobDet)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	response := ResponseGeneral{
		ResponseCode: http.StatusOK,
		RespMessage:  "Job Details successfully",
	}
	jsonResp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)

}

func GetJobsList(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting all jobs..")
	var jobs []JobDetails

	w.Header().Set("Content-Type", "application/json")

	result := DB.Find(&jobs)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			ResponseCode: http.StatusInternalServerError,
			RespMessage:  result.Error.Error(),
		})
		return
	}

	responseReturn := Response{
		ResponseCode: http.StatusOK,
		RespMessage:  "Jobs list fetched successfully",
		Data: JobsListData{
			JobsList: jobs,
		},
	}

	w.WriteHeader(http.StatusOK)
	jsonResp, err := json.Marshal(responseReturn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(jsonResp)

}

func isAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("isAuth running..")

	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := cookie.Value

	claims := &claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	UserName := claims.UserName

	fmt.Println("Username: ", UserName)

	if UserName != "" {
		var response Response
		response.ResponseCode = 200
		response.RespMessage = "Authorization Successful"
		response.Data = UserName

		jsonResp, _ := json.Marshal(response)

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	return

}
