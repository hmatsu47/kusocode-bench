package dbaccess

import (
	"fmt"

	"github.com/hmatsu47/kusocode-bench/dbmodel"
)

func FetchTeam(ipAddress string) (dbmodel.Team, error) {
	var t dbmodel.Team
	db := Connect()
	err := db.QueryRow("SELECT id, name, score, exec_flag FROM team WHERE ip_address = ?", ipAddress).Scan(&t.Id, &t.Name, &t.Score, &t.ExecFlag)
	if err != nil {
		fmt.Println(err)
	}
	return t, err
}

func UpdateTeam(team dbmodel.Team) error {
	db := Connect()
	upd, err := db.Prepare("UPDATE team SET score = ?, exec_flag = ? WHERE id = ?")
	if err != nil {
		fmt.Println(err)
	}
	upd.Exec(team.Score, team.ExecFlag, team.Id)
	return err
}
