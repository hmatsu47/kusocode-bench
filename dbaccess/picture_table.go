package dbaccess

import (
	"fmt"

	"github.com/hmatsu47/kusocode-bench/dbmodel"
)

func FetchPicturesAll() ([25]dbmodel.Picture, error) {
	var parray [25]dbmodel.Picture
	db := Connect()
	rs, err := db.Query("SELECT id, title, description, image FROM picture ORDER BY id")
	if err != nil {
		fmt.Println(err)
		return parray, err
	}
	var i = 0
	for rs.Next() {
		if i == 26 {
			return parray, err
		}
		p := dbmodel.Picture{}
		err = rs.Scan(&p.Id, &p.Title, &p.Description, &p.Image)
		if err != nil {
			fmt.Println(err)
			return parray, err
		}
		parray[i] = p
		i++
	}
	return parray, err
}
