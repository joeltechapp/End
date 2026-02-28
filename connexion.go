package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UserInfos struct {
	// PARTI IV
	DbUserNameIv  string `db:"db_user_name_iv"`
	DbPasswordIv  string `db:"db_password_iv"`
	DbHostIv      string `db:"db_host_iv"`
	DbPortIv      string `db:"db_port_iv"`
	DbNameIv      string `db:"db_name_iv"`
	DbTableNameIv string `db:"db_table_name_iv"`
	// PARTI DATA
	DbUserNameData  string `db:"db_user_name_data"`
	DbPasswordData  string `db:"db_password_data"`
	DbHostData      string `db:"db_host_data"`
	DbPortData      string `db:"db_port_data"`
	DbNameData      string `db:"db_name_data"`
	DbTableNameData string `db:"db_table_name_data"`
	// PARTI TAG
	DbUserNameTag  string `db:"db_user_name_tag"`
	DbPasswordTag  string `db:"db_password_tag"`
	DbHostTag      string `db:"db_host_tag"`
	DbPortTag      string `db:"db_port_tag"`
	DbNameTag      string `db:"db_name_tag"`
	DbTableNameTag string `db:"db_table_name_tag"`
	// ACTIVATION
	IsActive   int       `db:"is_active"`
	DateLimite time.Time `db:"date_limite"`
}

type contextKey string

const userInfosKey contextKey = "userInfos"

// --------------------
// Connexion MySQL normale
// --------------------
func dbConnec(user, password, host, port, dbname string) (*sql.DB,error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=skip-verify", user, password, host, port, dbname)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Erreur lors de la connexion MySQL :", err)
		return nil,err
	}
	if err=sqlDB.Ping();err!=nil {
		fmt.Println("Impossible de me connecter: ",err)
		return nil,err
	}
	return sqlDB,err
}

// --------------------
// Middleware
// --------------------
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FONT_URL"))

		cookie := cookieRecuperation(w, r, "user_id")
		if cookie == "nil" {
			json.NewEncoder(w).Encode(map[string]any{
				"login": "Veuillez vous inscrire ou vous connecter pour continuer",
			})
			return
		}

		// Connexion DB principale pour récupérer infos utilisateur
		sqlDB,err := dbConnec(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))
		if err!=nil {
			log.Println("connection impossible")
			return 
		}
		if sqlDB == nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}
		defer sqlDB.Close()

		var value UserInfos
		data, err := sqlDB.Query(`
			SELECT db_user_name_iv,db_password_iv,db_host_iv,db_port_iv,db_name_iv,db_table_name_iv,
			       db_user_name_data,db_password_data,db_host_data,db_port_data,db_name_data,db_table_name_data,
			       db_user_name_tag,db_password_tag,db_host_tag,db_port_tag,db_name_tag,db_table_name_tag,
			       is_active,date_limite
			FROM users
			WHERE id=?`, cookie)
		if err != nil {
			fmt.Println("Erreur récupération utilisateur :", err)
			return
		}
		defer data.Close()

		for data.Next() {
			err := data.Scan(
				&value.DbUserNameIv, &value.DbPasswordIv, &value.DbHostIv, &value.DbPortIv, &value.DbNameIv, &value.DbTableNameIv,
				&value.DbUserNameData, &value.DbPasswordData, &value.DbHostData, &value.DbPortData, &value.DbNameData, &value.DbTableNameData,
				&value.DbUserNameTag, &value.DbPasswordTag, &value.DbHostTag, &value.DbPortTag, &value.DbNameTag, &value.DbTableNameTag,
				&value.IsActive, &value.DateLimite)
			if err != nil {
				fmt.Println("Erreur scan :", err)
			}
		}
		if err:=data.Err();err!=nil {
			log.Println(err)
		}
		// Vérification abonnement
		_, err = sqlDB.Exec(`
			UPDATE users
			SET is_active = IF(date_limite >= NOW(),1,0) WHERE id=?
		`,cookie)
		if err != nil {
			fmt.Println("Erreur update :", err)
			return
		}
		if value.IsActive == 0 {
			json.NewEncoder(w).Encode(map[string]any{
				"Payment": "Vos données ne sont pas encore disponibles car aucun abonnement actif n'a été détecté. Une fois le paiement confirmé vos tableaux apparaîtront ici automatiquement.",
			})
			return
		}

		// Ajouter les infos dans le contexte
		ctx := context.WithValue(r.Context(), userInfosKey, value)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}