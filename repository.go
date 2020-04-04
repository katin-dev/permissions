package main

import "log"

func getUserPermissions(userId string) ([]string, error) {
	rows, err := db.Query(`
		SELECT p.name 
		FROM permission p
		JOIN role_permission rp ON rp.permission_id = p.id
		JOIN user_role ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1 	
	`, userId)

	if err != nil {
		log.Println("DB error: " + err.Error())
	}

	defer rows.Close()

	permissions := make([]string, 0)

	for rows.Next() {
		permission := ""
		rows.Scan(&permission)
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, error(err)
	}

	return permissions, nil
}
