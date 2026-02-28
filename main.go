package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// initTable s'assure que la table existe au lancement
func initTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(100),
		first_name VARCHAR(100),
		full_name VARCHAR(255),
		email VARCHAR(255) UNIQUE,
		password VARCHAR(255),
		db_user_name_iv VARCHAR(255),
		db_password_iv VARCHAR(255),
		db_host_iv VARCHAR(255),
		db_port_iv VARCHAR(255),
		db_name_iv VARCHAR(255),
		db_table_name_iv VARCHAR(255),
		db_user_name_data TEXT,
		db_password_data TEXT,
		db_host_data TEXT,
		db_port_data TEXT,
		db_name_data TEXT,
		db_table_name_data TEXT,
		db_user_name_tag VARCHAR(255),
		db_password_tag VARCHAR(255),
		db_host_tag VARCHAR(255),
		db_port_tag VARCHAR(255),
		db_name_tag VARCHAR(255),
		db_table_name_tag VARCHAR(255),
		is_active TINYINT(1) DEFAULT 0,
		date_limite DATETIME,
		created_at DATETIME
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Erreur d'initialisation de la table: %v", err)
	} else {
		fmt.Println("Vérification de la table 'users' terminée.")
	}
}

func main() {
	// 1. Chargement facultatif du .env
	godotenv.Load()

	// 2. Connexion à la DB principale pour l'initialisation
	// On utilise ta fonction dbConnec avec les variables d'environnement
	sqlDB, err := dbConnec(
		os.Getenv("DB_USER"), 
		os.Getenv("DB_PASSWORD"), 
		os.Getenv("DB_HOST"), 
		os.Getenv("DB_PORT"), 
		os.Getenv("DB_DATABASE"),
	)

	if err != nil {
		log.Println("Connexion impossible pour l'initialisation, le serveur continue...")
	} else {
		initTable(sqlDB)
		sqlDB.Close() // On ferme cette connexion temporaire
	}

	// 3. Configuration du Router Chi
	r := chi.NewRouter()

	r.Post("/users/inscription", inscription)
	r.Post("/users/connexion", connexion)
	r.Post("/users/admins/connexion", connexionAdmin)
	r.Post("/users/update", update)

	r.Route("/protected", func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Get("/data", dataOnly)
		r.Get("/colonnes", recuperationColonnes)
		r.Get("/data/tables", receiveTable)
		r.Get("/verifyDate", verifyLimitedate)
	})

	r.With(AdminMiddleware).Get("/protected/data/admin", dataAmins)
	r.With(AdminMiddleware).Post("/protected/data/acivate", active)
	
	// 4. Lancement du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 SaaS opérationnel sur le port %s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Erreur fatale serveur: %v", err)
	}
}
