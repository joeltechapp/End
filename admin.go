package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	    w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FONT_URL"))	

		cookieDec := cookieRecuperation(w, r, "Autoriser")

		if cookieDec == os.Getenv("MOT_DE_PASSE") {
			next.ServeHTTP(w, r)
			return 
		}
		http.Redirect(w, r,os.Getenv("FONT_URL") + "/admins/connexion", http.StatusFound)
	})
}

func dataAmins(w http.ResponseWriter, r *http.Request) {
	sql,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
	if err != nil {
		fmt.Println("Erreur lors de la conection ou im")
	}
	row, err := sql.Query("SELECT id,full_name,email,is_active FROM users")
	if err != nil {
		log.Printf("Erreur lors de la recuperation : %s",err)
		return
	}
	defer row.Close()
	
	colonnes, err := row.Columns()
	if err != nil {
		return
	}
	nombreColonne := len(colonnes)
	destinations := make([]any, nombreColonne)
	for i := range destinations {
		destinations[i] = new(any)
	}
	var resultats []map[string]any
	for row.Next() {
		err := row.Scan(destinations...)
		if err != nil {
			return
		}
		ligne := make(map[string]any)
		for i, colName := range colonnes {
			valeur := *(destinations[i].(*any))
			if b, ok := valeur.([]byte); ok {
				ligne[colName] = string(b)
			} else {
				ligne[colName] = valeur
			}
		}
		resultats = append(resultats, ligne)
	}
	if err := row.Err(); err != nil {
		return
	}
	json.NewEncoder(w).Encode(resultats)
}

func connexionAdmin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.FormValue("email")
	password := r.FormValue("password")
	emailVerify := strings.ToLower(os.Getenv("EMAIL"))

	if email == emailVerify {
		if password == os.Getenv("PASSWORD") {
			cookie := &http.Cookie{
				Name:     "Autoriser",
				Path:     "/",
				Value:    os.Getenv("MOT_DE_PASSE"),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   600,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, os.Getenv("FONT_URL")+"/admins/controller", http.StatusFound)
			return
		}
		http.Redirect(w, r, os.Getenv("FONT_URL")+"/admins/connexion", http.StatusFound)
		return
	}
	http.Redirect(w, r, os.Getenv("FONT_URL")+"/admins/connexion", http.StatusFound)
}

func active(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()
	id := r.FormValue("user_id")

	dateLimite := time.Now().Add(30*24 * time.Hour)
	isActive := 1
	sql,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))

	if err != nil {
		fmt.Println("Erreur lors de la connection")
		return
	}

	res, err := sql.Exec("UPDATE users SET is_active=?,date_limite=? WHERE id=?",isActive,dateLimite,id)
	if err != nil {
		fmt.Println("Erreur lors de la modification : ",err)
		return
	}
	_, err = res.RowsAffected()

	if err != nil {
		fmt.Println("Erreur lors de la modification : ", err)
		return 
	}

	http.Redirect(w, r, os.Getenv("FONT_URL")+"/admins/controller", http.StatusFound)
}}