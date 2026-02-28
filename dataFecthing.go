package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//-----------------------------TOUTES LES FONCTIONS A OPTIMISER/AMELIORER ------------------------------------

func useConnect(r *http.Request) (*sql.DB) {
	user := r.Context().Value(userInfosKey)
	dataUser := user.(UserInfos)

	userName, err := decrypt(dataUser.DbUserNameIv,dataUser.DbUserNameData,dataUser.DbUserNameTag)
	if err != nil {
		fmt.Println("erreur recuperation du nom")
		return nil
	}

	userPassword, err := decrypt(dataUser.DbPasswordIv,dataUser.DbPasswordData,dataUser.DbPasswordTag)
	if err != nil {
		fmt.Println("erreur recuperation")
		return nil
	}

	userHost, err := decrypt(dataUser.DbHostIv,dataUser.DbHostData,dataUser.DbHostTag)
	if err != nil {
		fmt.Println("erreur recuperation")
		return nil
	}

	userPort, err := decrypt(dataUser.DbPortIv,dataUser.DbPortData,dataUser.DbPortTag)
	if err != nil {
		fmt.Println("erreur recuperation")
		return nil
	}

	userDbName, err := decrypt(dataUser.DbNameIv,dataUser.DbNameData,dataUser.DbNameTag)
	if err != nil {
		fmt.Println("erreur recuperation")
		return nil
	}

	sql,err := dbConnec(userName,userPassword,userHost,userPort,userDbName)
	if err != nil {
		fmt.Println("Impossible se s'y connecter")
	}

	return sql
}

//FONCTION POUR LES TABLES

func receiveTable(w http.ResponseWriter, r *http.Request){
	sql:= useConnect(r)

	quary := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
	`
	rows, err := sql.Query(quary)
	if err != nil {
		return
	}	
	defer sql.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			fmt.Println("Erreur: ",err)
		}
		tables = append(tables, table)
	}
	json.NewEncoder(w).Encode(tables)
}

//FONTION POUR LES DONNEES DES TABLES

func dataOnly(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userInfosKey)
	dataUser := user.(UserInfos)

	userTableName, err := decrypt(dataUser.DbTableNameIv,dataUser.DbTableNameData,dataUser.DbTableNameTag)
	if err != nil {
		fmt.Println("erreur recuperation")
		return
	}

	sql := useConnect(r)   
	rows, err := sql.Query("SELECT * FROM " + userTableName)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"Erreur":"Cette la table n'existe pas dans votre base de données veuillez choisir une autre",
		})
		fmt.Printf("Erreur lors de la recuperation des données de la table %s",userTableName)
		return
	}	
	defer sql.Close()

	colonnes, err := rows.Columns()
	if err != nil {
		return
	}
	nombreColonne := len(colonnes)
	destinations := make([]any, nombreColonne)
	for i := range destinations {
		destinations[i] = new(any)
	}
	var resultats []map[string]any
	for rows.Next() {
		err := rows.Scan(destinations...)
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
	if err := rows.Err(); err != nil {
		fmt.Println("Oui")
		return
	}
	json.NewEncoder(w).Encode(resultats)
}

func recuperationColonnes(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userInfosKey)
	dataUser := user.(UserInfos)
	
	userTableName, err := decrypt(dataUser.DbTableNameIv,dataUser.DbTableNameData,dataUser.DbTableNameTag)
	if err != nil {
		fmt.Println("erreur du decodage")
		return
	}

	sql := useConnect(r)   
	rows, err := sql.Query("SELECT * FROM " + userTableName)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{
			"Erreur":"Cette la table n'existe pas dans votre base de données veuillez choisir une autre",
		})
		fmt.Printf("Erreur lors de la recuperation des données de la table %s",userTableName)
		return
	}
	defer sql.Close()

	colonnes, err := rows.Columns()
	if err != nil {
		return
	}
	json.NewEncoder(w).Encode(colonnes)
}

// func mask(name string) string {
// 	runes := []rune(name)
// 	lenght := len(runes)

// 	first := string(runes[0])
// 	last := string(runes[lenght-1])

// 	middle := strings.Repeat("*",lenght-2)

// 	return first + middle + last
// }

// func currentValue(w http.ResponseWriter, r *http.Request){
// 	user := r.Context().Value(userInfosKey)
// 	dataUser := user.(UserInfos)

// 	userName, err := decrypt(dataUser.DbUserNameIv,dataUser.DbUserNameData,dataUser.DbUserNameTag)
// 	if err != nil {
// 		fmt.Println("erreur recuperation du nom")
// 		return 
// 	}

// 	userPassword, err := decrypt(dataUser.DbPasswordIv,dataUser.DbPasswordData,dataUser.DbPasswordTag)
// 	if err != nil {
// 		fmt.Println("erreur recuperation")
// 		return 
// 	}

// 	userHost, err := decrypt(dataUser.DbHostIv,dataUser.DbHostData,dataUser.DbHostTag)
// 	if err != nil {
// 		fmt.Println("erreur recuperation")
// 		return
// 	}

// 	userPort, err := decrypt(dataUser.DbPortIv,dataUser.DbPortData,dataUser.DbPortTag)
// 	if err != nil {
// 		fmt.Println("erreur recuperation")
// 		return 
// 	}

// 	userDbName, err := decrypt(dataUser.DbNameIv,dataUser.DbNameData,dataUser.DbNameTag)
// 	if err != nil {
// 		fmt.Println("erreur recuperation")
// 		return 
// 	}

// 	userTableName, err := decrypt(dataUser.DbTableNameIv,dataUser.DbTableNameData,dataUser.DbTableNameTag)
// 	if err != nil {
// 		fmt.Println("erreur du decodage")
// 		return
// 	}

// 	maskName := mask(userName)
// 	maskPassword := mask(userPassword)
// 	maskHost := mask(userHost)
// 	maskPort := mask(userPort)
// 	maskDbName := mask(userDbName)
// 	maskTableName := mask(userTableName)

// 	current := []string{
// 		maskName,
// 		maskPassword,
// 		maskHost,
// 		maskPort,
// 		maskDbName,
// 		maskTableName,
// 	}
// 	json.NewEncoder(w).Encode(current)
// }


func verifyLimitedate(w http.ResponseWriter, r *http.Request){
	user := r.Context().Value(userInfosKey)
	dataUser := user.(UserInfos)
	
	now := time.Now()
	dateLimite := dataUser.DateLimite
	isActive := dataUser.IsActive
	
	userTableName, err := decrypt(dataUser.DbTableNameIv,dataUser.DbTableNameData,dataUser.DbTableNameTag)
	if err != nil {
		fmt.Println("erreur du decodage")
		return
	}
	
	if dateLimite.Before(now) {
		json.NewEncoder(w).Encode(map[string]any{
			"is_inactive":isActive,
		})
		return
	}
	formate := dateLimite.Local().Format("02 January 2006 à 15:04 ")
	json.NewEncoder(w).Encode(map[string]any{
		"date_limite":formate,
		"is_active":isActive,
		"current_table":userTableName,
	})
}