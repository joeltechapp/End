package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string)(string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash),nil
}

func cookieRecuperation(_ http.ResponseWriter, r *http.Request,name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "nil"
	}
	return cookie.Value
}

func inscription(w http.ResponseWriter, r *http.Request){

	sql,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"dbEtat":"Connection impossible à la base de données",
		})
		return 
	}
	r.ParseForm()

	name := r.FormValue("name")
	firstName := r.FormValue("firstName")
	fullName := name + " " + firstName
	emailRecive := r.FormValue("email")
	email := strings.ToLower(emailRecive)
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("passwordConfirm")

	if password != passwordConfirm {
		http.Redirect(w,r,os.Getenv("FONT_URL")+"/inscription?error=403",http.StatusFound)
		return
	}
	dbHost := r.FormValue("dbHost")
	dbPort := r.FormValue("dbPort")
	dbName := r.FormValue("dbName")
	dbUserName := r.FormValue("dbUserName")
	dbPassword := r.FormValue("dbPassword")
	DbTableNameData := r.FormValue("dbdbTableName")

	encryptHost, err := encrypt(dbHost)
	encryptPort, err := encrypt(dbPort)
	encryptName, err := encrypt(dbName)
	encryptTableName, err := encrypt(DbTableNameData)
	encryptUserName, err := encrypt(dbUserName)
	encryptPassword, err := encrypt(dbPassword)

	dateLimite := time.Now().Add(7*24*time.Hour)
	isActive := 1

	if err != nil {
		fmt.Println("erreur lors du cryptage")
		return
	}

	
	idByte := make([]byte,32)
	_, err = rand.Read(idByte)
	if err!=nil {
		fmt.Println("ERREUR lors de la generation")
		return
	}
	id := hex.EncodeToString(idByte) 
	
	hash, err := hashPassword(password)
	if err != nil {
		fmt.Println("Erreur lors du hashage")
		return
	}

	_, err = sql.Exec("INSERT INTO users (id,name,first_name,full_name,email,password,db_user_name_iv,db_password_iv,db_host_iv,db_port_iv,db_name_iv,db_table_name_iv, db_user_name_data,db_password_data,db_host_data,db_port_data,db_name_data,db_table_name_data, db_user_name_tag,db_password_tag,db_host_tag,db_port_tag,db_name_tag,db_table_name_tag,is_active,date_limite) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",id,name,firstName,fullName,email,hash,encryptUserName.IV,encryptPassword.IV,encryptHost.IV,encryptPort.IV,encryptName.IV,encryptTableName.IV,encryptUserName.Data,encryptPassword.Data,encryptHost.Data,encryptPort.Data,encryptName.Data,encryptTableName.Data,encryptUserName.Tag,encryptPassword.Tag,encryptHost.Tag,encryptPort.Tag,encryptName.Tag,encryptTableName.Tag,isActive,dateLimite)
	if err!=nil {
		log.Printf("erreur lors de l'insertion %s de l'email=%s",err,email)
		http.Redirect(w,r,os.Getenv("FONT_URL")+"/inscription?error=403",http.StatusFound)
		return
	}

	cookie := &http.Cookie{
		Name: "user_id",
		Path: "/",
		Value: id,
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge: 2592000,
	}
	
	http.SetCookie(w,cookie)
	http.Redirect(w,r,os.Getenv("FONT_URL")+"/dashboard",http.StatusFound)
}


var (
	// Map pour limiter les tentatives
	failedLogin = make(map[string]int)
	mutex       = sync.Mutex{}
	MAX_TRIES   = 5
)

func connexion(w http.ResponseWriter, r *http.Request) {
	sql,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))

	if err!= nil {
		fmt.Println("Impossible de se connecter")
		return
	}

	r.ParseForm()

	emailRecive := r.FormValue("email")
	email := strings.ToLower(emailRecive)
	password := r.FormValue("password")

	// Vérifier les tentatives
	mutex.Lock()
	if val, exists := failedLogin[email]; exists && val <= 0 {
		mutex.Unlock()
		http.Redirect(w, r, os.Getenv("FONT_URL")+"/connexion?error=403", http.StatusFound)
		return
	}
	mutex.Unlock()

	var (
		hash   string
		userId string
	)

	err = sql.QueryRow("SELECT id,password FROM users WHERE email=?", email).Scan(&userId, &hash)
	if err != nil {
		log.Printf("Erreur l'email=%s n'existe pas : %s", email, err)
		updateFailedLogin(email)
		http.Redirect(w, r, os.Getenv("FONT_URL")+"/connexion?error=401", http.StatusFound)
		return
	}
	defer sql.Close()

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("l'utilisateur avec l'email=%s n'as pas inscrit le mot de passe correcte : %s", email, err)
		updateFailedLogin(email)
		http.Redirect(w, r, os.Getenv("FONT_URL")+"/connexion?error=401Unautorized", http.StatusFound)
		return
	}

	// Réinitialiser compteur si succès
	resetFailedLogin(email)

	cookie := &http.Cookie{
		Name:     "user_id",
		Path:     "/",
		Value:    userId,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   2592000,
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, os.Getenv("FONT_URL")+"/dashboard", http.StatusFound)
}

// Décrémente ou initialise le compteur et programme reset après 24h
func updateFailedLogin(email string) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, exists := failedLogin[email]; !exists {
		failedLogin[email] = MAX_TRIES - 1
	} else {
		failedLogin[email]--
	}

	// Reset automatique après 24h
	go func(email string) {
		time.Sleep(24 * time.Hour)
		resetFailedLogin(email)
	}(email)
}

func resetFailedLogin(email string) {
	mutex.Lock()
	defer mutex.Unlock()
	failedLogin[email] = MAX_TRIES
}

func update(w http.ResponseWriter, r *http.Request){

	sql,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"dbEtat":"Connection impossible à la base de données",
		})
		return 
	}

	r.ParseForm()

	userId := cookieRecuperation(w,r,"user_id")

	table := r.FormValue("tableName")

	tableEncrypt, err := encrypt(table)
	if err != nil {
		fmt.Println("Errer lors de l'encryptage")
		return
	}

	res, err := sql.Exec("UPDATE users SET db_table_name_iv=?,db_table_name_data=?,db_table_name_tag=? WHERE id=?",tableEncrypt.IV,tableEncrypt.Data,tableEncrypt.Tag,userId)

	rows, _ := res.RowsAffected()

	if rows == 0 {
		http.Error(w,"user not fund",404)
		return
	}
	http.Redirect(w,r,os.Getenv("FONT_URL")+"/dashboard",http.StatusFound)
}
